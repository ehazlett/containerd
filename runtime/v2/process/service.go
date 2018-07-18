/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless ruired by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package process

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	eventstypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/api/types/task"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/runtime"
	"github.com/containerd/containerd/runtime/v2/shim"
	taskapi "github.com/containerd/containerd/runtime/v2/task"
	"github.com/containerd/fifo"
	ptypes "github.com/gogo/protobuf/types"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var _ = (taskapi.TaskService)(&Service{})

var (
	ErrNotSupported = errors.New("operation not supported")
	empty           = &ptypes.Empty{}
)

type command struct {
	*exec.Cmd
	stdin  string
	stdout string
	stderr string
	// set on Start
	pio *pipeIO
}

type commandStart struct {
	id  string
	pid int
}

type commandEvent struct {
	id     string
	pid    int
	status uint32
	exited time.Time
	err    error
}

type Service struct {
	id      string
	context context.Context
	events  chan interface{}
	exitCh  chan *commandEvent
	mu      sync.Mutex

	bundle   string
	commands map[string]*command
	pids     map[string]int
	pidCh    chan *commandStart
	errCh    chan error
}

var (
	logPath = filepath.Join("/tmp", "shim.log")
)

func New(ctx context.Context, id string, publisher events.Publisher) (shim.Shim, error) {
	s := &Service{
		id:       id,
		context:  ctx,
		events:   make(chan interface{}, 128),
		exitCh:   make(chan *commandEvent),
		commands: make(map[string]*command),
		pids:     make(map[string]int),
		pidCh:    make(chan *commandStart),
		errCh:    make(chan error),
	}
	go s.forward(publisher)
	return s, nil
}

func (s *Service) logMsg(msg string) {
	fmt.Println(msg)
}

func newCommand(ctx context.Context, containerdBinary, containerdAddress string) (*exec.Cmd, error) {
	os.Remove(logPath)
	logfile, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}
	ns, err := namespaces.NamespaceRequired(ctx)
	if err != nil {
		return nil, err
	}
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	args := []string{
		"-namespace", ns,
		"-address", containerdAddress,
		"-publish-binary", containerdBinary,
	}
	cmd := exec.Command(self, args...)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "GOMAXPROCS=2")
	cmd.Stdout = logfile
	//cmd.Stderr = logfile
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return cmd, nil
}

func (s *Service) StartShim(ctx context.Context, id, containerdBinary, containerdAddress string) (string, error) {
	cmd, err := newCommand(ctx, containerdBinary, containerdAddress)
	if err != nil {
		return "", err
	}
	address, err := shim.AbstractAddress(ctx, id)
	if err != nil {
		return "", err
	}
	socket, err := shim.NewSocket(address)
	if err != nil {
		return "", err
	}
	defer socket.Close()
	f, err := socket.File()
	if err != nil {
		return "", err
	}
	defer f.Close()

	cmd.ExtraFiles = append(cmd.ExtraFiles, f)

	if err := cmd.Start(); err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			cmd.Process.Kill()
		}
	}()
	// make sure to wait after start
	go cmd.Wait()
	if err := shim.WritePidFile("shim.pid", cmd.Process.Pid); err != nil {
		return "", err
	}
	if err := shim.WriteAddress("address", address); err != nil {
		return "", err
	}
	if err := shim.SetScore(cmd.Process.Pid); err != nil {
		return "", errors.Wrap(err, "failed to set OOM Score on shim")
	}
	go s.errorHandler()
	go s.processExits()
	return address, nil
}

func (s *Service) Create(ctx context.Context, r *taskapi.CreateTaskRequest) (*taskapi.CreateTaskResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	//var opts RuntimeOptions
	//if r.Options != nil {
	//	v, err := typeurl.UnmarshalAny(r.Options)
	//	if err != nil {
	//		return nil, err
	//	}
	//	opts = *v.(*RuntimeOptions)
	//}

	s.bundle = r.Bundle
	s.id = r.ID

	s.logMsg(fmt.Sprintf("rootfs: %+v", r.Rootfs))

	var spec *specs.Spec
	data, err := ioutil.ReadFile(filepath.Join(r.Bundle, "config.json"))
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, errdefs.ToGRPC(err)
	}

	if len(spec.Process.Args) == 0 {
		return nil, errdefs.ToGRPC(fmt.Errorf("no process args specified"))
	}

	cmd := spec.Process.Args[0]
	args := []string{}
	if len(spec.Process.Args) > 1 {
		args = spec.Process.Args[1:]
	}

	c := exec.Command(cmd, args...)
	c.Dir = spec.Process.Cwd
	c.Env = spec.Process.Env
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	sc := &command{
		c,
		r.Stdin,
		r.Stdout,
		r.Stderr,
		nil,
	}
	s.commands[r.ID] = sc

	return &taskapi.CreateTaskResponse{
		Pid: 0,
	}, nil
}

func (s *Service) Start(ctx context.Context, r *taskapi.StartRequest) (*taskapi.StartResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := s.commands[r.ID]
	if c == nil {
		return nil, errdefs.ToGRPCf(errdefs.ErrNotFound, "start: command %s", r.ID)
	}

	s.logMsg("setting up IO")
	cwg := &sync.WaitGroup{}
	cwg.Add(1)
	var ioErr error
	go func() {
		if err := s.setupIO(ctx, r.ID, c, cwg); err != nil {
			ioErr = err
		}
	}()
	cwg.Wait()
	if ioErr != nil {
		return nil, errdefs.ToGRPC(ioErr)
	}

	s.logMsg("starting process")
	if err := c.Start(); err != nil {
		return nil, errdefs.ToGRPC(err)
	}

	s.logMsg(fmt.Sprintf("process started: pid=%d", c.Process.Pid))

	s.pids[r.ID] = c.Process.Pid
	if err := s.monitor(r.ID, c.Process.Pid); err != nil {
		return nil, errdefs.ToGRPC(err)
	}

	s.events <- &eventstypes.TaskCreate{
		ContainerID: s.id,
		Bundle:      s.bundle,
		Pid:         uint32(c.Process.Pid),
	}

	s.logMsg("event published TaskCreate")

	return &taskapi.StartResponse{
		Pid: uint32(c.Process.Pid),
	}, nil
}

func (s *Service) State(ctx context.Context, r *taskapi.StateRequest) (*taskapi.StateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pid, _ := s.pids[r.ID]
	//if !ok {
	//	return nil, errdefs.ToGRPCf(errdefs.ErrNotFound, "state: command %s", r.ID)
	//}

	resp := &taskapi.StateResponse{
		ID:     r.ID,
		Bundle: s.bundle,
		Pid:    uint32(pid),
	}

	exitCode := uint32(0)
	status := task.StatusStopped
	if pid != 0 {
		p, _ := os.FindProcess(pid)
		if p != nil {
			statusErr := syscall.Kill(pid, syscall.Signal(0))
			if statusErr == nil {
				status = task.StatusRunning
			}
		}
	}

	resp.ExitStatus = exitCode
	resp.Status = status

	return resp, nil
}

func (s *Service) Delete(ctx context.Context, r *taskapi.DeleteRequest) (*taskapi.DeleteResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pid, ok := s.pids[r.ID]
	if ok {
		s.logMsg(fmt.Sprintf("delete: finding pid=%d", pid))
		p, err := os.FindProcess(pid)
		if err != nil {
			return nil, errors.Wrapf(errdefs.ErrNotFound, "delete: pid %d", pid)
		}
		s.logMsg(fmt.Sprintf("delete: killing pid=%d", pid))
		statusErr := syscall.Kill(pid, syscall.Signal(0))
		if statusErr == nil {
			if err := p.Kill(); err != nil {
				return nil, errdefs.ToGRPC(err)
			}
		}
		delete(s.pids, r.ID)
	}
	if _, ok := s.commands[r.ID]; ok {
		//s.logMsg("closing pio")
		//if c.pio != nil {
		//	c.pio.Close()
		//}
		delete(s.commands, r.ID)
	}

	if r.ID == s.id {
		s.logMsg(fmt.Sprintf("delete: removing bundle %s", s.bundle))
		os.RemoveAll(s.bundle)
		close(s.exitCh)
	}

	s.logMsg("delete: complete")

	return &taskapi.DeleteResponse{
		ExitedAt: time.Now(),
		Pid:      uint32(pid),
	}, nil
}

func (s *Service) Pids(ctx context.Context, r *taskapi.PidsRequest) (*taskapi.PidsResponse, error) {
	info := []*task.ProcessInfo{}
	for _, pid := range s.pids {
		info = append(info, &task.ProcessInfo{
			Pid: uint32(pid),
		})
	}
	return &taskapi.PidsResponse{
		Processes: info,
	}, nil
}

func (s *Service) Pause(ctx context.Context, _ *taskapi.PauseRequest) (*ptypes.Empty, error) {
	return empty, errdefs.ToGRPC(ErrNotSupported)
}

func (s *Service) Resume(ctx context.Context, _ *taskapi.ResumeRequest) (*ptypes.Empty, error) {
	return empty, errdefs.ToGRPC(ErrNotSupported)
}

func (s *Service) Checkpoint(ctx context.Context, r *taskapi.CheckpointTaskRequest) (*ptypes.Empty, error) {
	return empty, errdefs.ToGRPC(ErrNotSupported)
}

func (s *Service) Kill(ctx context.Context, r *taskapi.KillRequest) (*ptypes.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sig := syscall.Signal(r.Signal)
	if r.All {
		for _, pid := range s.pids {
			p, err := os.FindProcess(pid)
			if err != nil {
				logrus.Warnf("unable to find process %d", pid)
				continue
			}
			if err := p.Signal(sig); err != nil {
				return empty, errdefs.ToGRPC(err)
			}
		}
	} else {
		id := r.ID
		if id == "" {
			id = s.id
		}
		pid, ok := s.pids[id]
		if ok {
			p, err := os.FindProcess(pid)
			if err != nil {
				return empty, errdefs.ToGRPC(err)
			}
			if err := p.Signal(sig); err != nil {
				return empty, errdefs.ToGRPC(err)
			}
		}
	}
	return empty, nil
}

func (s *Service) Exec(ctx context.Context, r *taskapi.ExecProcessRequest) (*ptypes.Empty, error) {
	return empty, nil
}

func (s *Service) ResizePty(ctx context.Context, r *taskapi.ResizePtyRequest) (*ptypes.Empty, error) {
	return empty, nil
}

func (s *Service) CloseIO(ctx context.Context, r *taskapi.CloseIORequest) (*ptypes.Empty, error) {
	return empty, errdefs.ErrNotImplemented
}

func (s *Service) Update(ctx context.Context, r *taskapi.UpdateTaskRequest) (*ptypes.Empty, error) {
	return empty, nil
}

func (s *Service) Wait(ctx context.Context, r *taskapi.WaitRequest) (*taskapi.WaitResponse, error) {
	if _, ok := s.commands[r.ID]; !ok {
		return nil, errdefs.ToGRPCf(errdefs.ErrNotFound, "wait: command %s", s.id)
	}

	s.logMsg("waiting for command exit event")

	evt := <-s.exitCh

	s.logMsg(fmt.Sprintf("wait complete: %+v", evt))

	return &taskapi.WaitResponse{
		ExitStatus: evt.status,
		ExitedAt:   evt.exited,
	}, nil
}

func (s *Service) Stats(ctx context.Context, r *taskapi.StatsRequest) (*taskapi.StatsResponse, error) {
	return &taskapi.StatsResponse{}, nil
}

// Connect returns shim information such as the shim's pid
func (s *Service) Connect(ctx context.Context, r *taskapi.ConnectRequest) (*taskapi.ConnectResponse, error) {
	pid, ok := s.pids[s.id]
	if !ok {
		return nil, errdefs.ToGRPCf(errdefs.ErrNotFound, "command %s", s.id)
	}
	return &taskapi.ConnectResponse{
		ShimPid: uint32(os.Getpid()),
		TaskPid: uint32(pid),
	}, nil
}

func (s *Service) Shutdown(ctx context.Context, r *taskapi.ShutdownRequest) (*ptypes.Empty, error) {
	s.logMsg("shutdown")
	os.Exit(0)
	return empty, nil
}

func (s *Service) Cleanup(ctx context.Context) (*taskapi.DeleteResponse, error) {
	s.logMsg(fmt.Sprintf("cleanup: removing bundle %s", s.bundle))
	_ = os.RemoveAll(s.bundle)
	return &taskapi.DeleteResponse{
		ExitedAt: time.Now(),
	}, nil
}

func (s *Service) forward(publisher events.Publisher) {
	for e := range s.events {
		if err := publisher.Publish(s.context, getTopic(s.context, e), e); err != nil {
			log.G(s.context).WithError(err).Error("post event")
		}
	}
}

func (s *Service) setupIO(ctx context.Context, id string, c *command, cwg *sync.WaitGroup) error {
	wg := &sync.WaitGroup{}
	pio, err := NewIO()
	if err != nil {
		return err
	}
	c.pio = pio
	s.logMsg("pio initialized")

	if c.stdin != "" {
		s.logMsg(fmt.Sprintf("configuring stdin: %s", c.stdin))
		c.Stdin = pio.in.r
		in, err := fifo.OpenFifo(ctx, c.stdin, syscall.O_RDONLY, 0)
		if err != nil {
			return err
		}
		go copyPipe(pio.Stdin(), in, wg)
		s.logMsg("stdin: started copyPipe")
	}
	if c.stdout != "" {
		s.logMsg(fmt.Sprintf("configuring stdout: %s", c.stdout))
		c.Stdout = pio.out.w
		outw, err := fifo.OpenFifo(ctx, c.stdout, syscall.O_WRONLY, 0)
		if err != nil {
			return err
		}
		go copyPipe(outw, pio.out.r, wg)
		s.logMsg("stdout: started copyPipe")
	}

	cwg.Done()
	wg.Wait()

	return nil
}

func (s *Service) errorHandler() {
	for err := range s.errCh {
		s.logMsg(fmt.Sprintf("shim error: %s", err))
	}
}

// monitor starts an exitHandler for a process upon start
func (s *Service) monitor(id string, pid int) error {
	s.logMsg(fmt.Sprintf("monitor: starting exit handler: id=%s pid=%d", id, pid))
	if err := s.exitHandler(id, pid); err != nil {
		return err
	}
	return nil
}

// exitHandler waits for the process to exit to publish exit events
func (s *Service) exitHandler(id string, pid int) error {
	s.logMsg(fmt.Sprintf("waitForExit: id=%s pid=%d", id, pid))
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	go func() {
		state, err := p.Wait()
		e, eErr := getExitCode(state)
		if err == nil {
			err = eErr
		}
		s.logMsg(fmt.Sprintf("process exit: id=%s pid=%d", id, pid))
		s.exitCh <- &commandEvent{
			id:     id,
			pid:    pid,
			status: e,
			exited: time.Now(),
			err:    err,
		}
		if c, ok := s.commands[id]; ok {
			s.logMsg("closing pio")
			if c.pio != nil {
				c.pio.Close()
			}
		}
		delete(s.pids, id)
		s.logMsg(fmt.Sprintf("published event: id=%s pid=%d", id, pid))
	}()
	return nil
}

func (s *Service) processExits() {
	for e := range s.exitCh {
		s.logMsg(fmt.Sprintf("exit event: %+v", e))
		s.checkCommand(e)
	}
}

func (s *Service) checkCommand(e *commandEvent) {
	for id, _ := range s.commands {
		s.logMsg(fmt.Sprintf("checkCommand: id=%s", id))
		if id == e.id {
			s.logMsg(fmt.Sprintf("publishing TaskExit id=%s", e.id))
			s.events <- &eventstypes.TaskExit{
				ContainerID: s.id,
				ID:          e.id,
				Pid:         uint32(e.pid),
				ExitStatus:  e.status,
				ExitedAt:    e.exited,
			}
			return
		}
	}
}

func copyPipe(dst io.WriteCloser, src io.ReadCloser, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		io.Copy(dst, src)
		src.Close()
		dst.Close()
		wg.Done()
	}()
}

func getTopic(ctx context.Context, e interface{}) string {
	switch e.(type) {
	case *eventstypes.TaskCreate:
		return runtime.TaskCreateEventTopic
	case *eventstypes.TaskStart:
		return runtime.TaskStartEventTopic
	case *eventstypes.TaskOOM:
		return runtime.TaskOOMEventTopic
	case *eventstypes.TaskExit:
		return runtime.TaskExitEventTopic
	case *eventstypes.TaskDelete:
		return runtime.TaskDeleteEventTopic
	case *eventstypes.TaskExecAdded:
		return runtime.TaskExecAddedEventTopic
	case *eventstypes.TaskExecStarted:
		return runtime.TaskExecStartedEventTopic
	case *eventstypes.TaskPaused:
		return runtime.TaskPausedEventTopic
	case *eventstypes.TaskResumed:
		return runtime.TaskResumedEventTopic
	case *eventstypes.TaskCheckpointed:
		return runtime.TaskCheckpointedEventTopic
	default:
		logrus.Warnf("no topic for type %#v", e)
	}
	return runtime.TaskUnknownTopic
}
