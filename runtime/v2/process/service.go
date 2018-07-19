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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	eventstypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/api/types/task"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/runtime"
	"github.com/containerd/containerd/runtime/v2/shim"
	taskapi "github.com/containerd/containerd/runtime/v2/task"
	runc "github.com/containerd/go-runc"
	ptypes "github.com/gogo/protobuf/types"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var _ = (shim.Shim)(&Service{})

var (
	ErrNotSupported = errors.New("operation not supported")
	empty           = &ptypes.Empty{}
)

type command struct {
	*exec.Cmd

	id     string
	stdin  string
	stdout string
	stderr string
	// set on Create()
	mounts []*mount.Mount
	// set on Start()
	pio        *pipeIO
	pid        int
	exit       chan struct{}
	exitedAt   time.Time
	exitStatus int
}

func (c *command) setExited(e runc.Exit) {
	c.exitedAt = e.Timestamp
	c.exitStatus = e.Status
	close(c.exit)
}

type Service struct {
	id      string
	context context.Context
	events  chan interface{}
	exits   chan runc.Exit
	mu      sync.Mutex

	bundle   string
	rootfs   string
	commands map[string]*command
	pid      int
}

var (
	logPath = filepath.Join("/tmp", "shim.log")
)

func New(ctx context.Context, id string, publisher events.Publisher) (shim.Shim, error) {
	s := &Service{
		id:       id,
		context:  ctx,
		events:   make(chan interface{}, 128),
		exits:    shim.Default.Subscribe(),
		commands: make(map[string]*command),
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
	cmd.Stderr = logfile
	cmd.SysProcAttr = getSysProcAttr()
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

	ec, err := shim.Default.Start(cmd)
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			cmd.Process.Kill()
			shim.Default.Wait(cmd, ec)
		}
	}()
	if err := shim.WritePidFile("shim.pid", cmd.Process.Pid); err != nil {
		return "", err
	}
	if err := shim.WriteAddress("address", address); err != nil {
		return "", err
	}
	if err := shim.SetScore(cmd.Process.Pid); err != nil {
		return "", errors.Wrap(err, "failed to set OOM Score on shim")
	}
	return address, nil
}

func (s *Service) Create(ctx context.Context, r *taskapi.CreateTaskRequest) (_ *taskapi.CreateTaskResponse, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	go s.processExits()

	rootfs := filepath.Join(r.Bundle, "rootfs")

	s.bundle = r.Bundle
	s.id = r.ID
	s.rootfs = rootfs

	s.logMsg(fmt.Sprintf("rootfs: %+v", r.Rootfs))

	var spec *specs.Spec
	data, err := ioutil.ReadFile(filepath.Join(r.Bundle, "config.json"))
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, errdefs.ToGRPC(err)
	}

	var mounts []*mount.Mount
	for _, m := range r.Rootfs {
		m := &mount.Mount{
			Type:    m.Type,
			Source:  m.Source,
			Options: m.Options,
		}
		if err := m.Mount(rootfs); err != nil {
			return nil, errors.Wrapf(err, "failed to mount rootfs component %v", m)
		}
		mounts = append(mounts, m)
	}
	defer func() {
		if err != nil {
			if _, err := s.Cleanup(ctx); err != nil {
				logrus.WithError(err).Warn("failed to cleanup")
			}
		}
	}()

	if len(spec.Process.Args) == 0 {
		return nil, errdefs.ToGRPC(fmt.Errorf("no process args specified"))
	}

	cmd := spec.Process.Args[0]
	args := []string{}
	if len(spec.Process.Args) > 1 {
		args = spec.Process.Args[1:]
	}

	if filepath.IsAbs(cmd) {
		cmd = filepath.Join(rootfs, cmd)
	} else {
		// check process env for path and search path for process
		c, err := resolveRootfsCommand(rootfs, cmd, spec.Process.Env)
		if err != nil {
			return nil, errdefs.ToGRPC(err)
		}
		cmd = c
	}

	s.logMsg(fmt.Sprintf("cmd: %s", cmd))

	c := exec.Command(cmd, args...)
	c.Dir = rootfs
	c.Env = spec.Process.Env
	c.SysProcAttr = getSysProcAttr()
	sc := &command{
		Cmd:    c,
		id:     r.ID,
		stdin:  r.Stdin,
		stdout: r.Stdout,
		stderr: r.Stderr,
		mounts: mounts,
		exit:   make(chan struct{}),
	}
	s.commands[r.ID] = sc

	return &taskapi.CreateTaskResponse{
		Pid: 0,
	}, nil
}

func (s *Service) Start(ctx context.Context, r *taskapi.StartRequest) (_ *taskapi.StartResponse, err error) {
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
		if err := setupIO(ctx, r.ID, c, cwg); err != nil {
			ioErr = err
		}
	}()
	cwg.Wait()
	if ioErr != nil {
		return nil, errdefs.ToGRPC(ioErr)
	}
	defer func() {
		if err != nil {
			if _, err := s.Cleanup(ctx); err != nil {
				logrus.WithError(err).Warn("failed to cleanup")
			}
		}
	}()

	s.logMsg("starting process")
	if _, err := shim.Default.Start(c.Cmd); err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	c.pio.out.w.Close()
	c.pio.err.w.Close()

	s.logMsg(fmt.Sprintf("process started: pid=%d", c.Process.Pid))

	c.pid = c.Process.Pid
	s.pid = c.Process.Pid

	return &taskapi.StartResponse{
		Pid: uint32(c.Process.Pid),
	}, nil
}

func (s *Service) State(ctx context.Context, r *taskapi.StateRequest) (*taskapi.StateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := s.commands[r.ID]
	if c == nil {
		return nil, errdefs.ToGRPCf(errdefs.ErrNotFound, "delete: command %s", r.ID)
	}
	pid := c.pid

	resp := &taskapi.StateResponse{
		ID:     r.ID,
		Bundle: s.bundle,
		Pid:    uint32(pid),
	}

	status := task.StatusStopped
	if pid != 0 {
		p, _ := os.FindProcess(pid)
		if p != nil {
			statusErr := processRunning(pid)
			if statusErr == nil {
				status = task.StatusRunning
			}
		}
	}

	resp.ExitStatus = uint32(c.exitStatus)
	resp.Status = status

	return resp, nil
}

func (s *Service) Delete(ctx context.Context, r *taskapi.DeleteRequest) (*taskapi.DeleteResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := s.commands[r.ID]
	if c == nil {
		return nil, errdefs.ToGRPCf(errdefs.ErrNotFound, "delete: command %s", r.ID)
	}
	pid := c.pid
	s.logMsg(fmt.Sprintf("delete: finding pid=%d", pid))
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil, errors.Wrapf(errdefs.ErrNotFound, "delete: pid %d", pid)
	}
	s.logMsg(fmt.Sprintf("delete: killing pid=%d", pid))
	statusErr := processRunning(pid)
	if statusErr == nil {
		if err := p.Kill(); err != nil {
			return nil, errdefs.ToGRPC(err)
		}
	}
	if c.pio != nil {
		c.pio.Close()
	}
	if err := mount.UnmountAll(s.rootfs, 0); err != nil {
		return nil, err
	}

	delete(s.commands, r.ID)

	s.logMsg("delete: complete")

	return &taskapi.DeleteResponse{
		ExitedAt: time.Now(),
		Pid:      uint32(c.pid),
	}, nil
}

func (s *Service) Pids(ctx context.Context, r *taskapi.PidsRequest) (*taskapi.PidsResponse, error) {
	info := []*task.ProcessInfo{}
	for _, c := range s.commands {
		info = append(info, &task.ProcessInfo{
			Pid: uint32(c.pid),
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
		for _, c := range s.commands {
			p, err := os.FindProcess(c.pid)
			if err != nil {
				logrus.Warnf("unable to find process %d", c.pid)
				continue
			}
			if err := p.Signal(sig); err != nil {
				return empty, errdefs.ToGRPC(err)
			}
		}
	} else {
		c := s.commands[r.ID]
		if c == nil {
			return nil, errdefs.ToGRPCf(errdefs.ErrNotFound, "kill: command %s", r.ID)
		}
		p, err := os.FindProcess(c.pid)
		if err != nil {
			return empty, errdefs.ToGRPC(err)
		}
		if err := p.Signal(sig); err != nil {
			return empty, errdefs.ToGRPC(err)
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
	if c, ok := s.commands[r.ID]; ok {
		if c.pio != nil {
			if err := c.pio.Close(); err != nil {
				return empty, errdefs.ToGRPC(err)
			}
		}
	}
	return empty, nil
}

func (s *Service) Update(ctx context.Context, r *taskapi.UpdateTaskRequest) (*ptypes.Empty, error) {
	return empty, nil
}

func (s *Service) Wait(ctx context.Context, r *taskapi.WaitRequest) (*taskapi.WaitResponse, error) {
	c, ok := s.commands[r.ID]
	if !ok {
		return nil, errdefs.ToGRPCf(errdefs.ErrNotFound, "wait: command %s", s.id)
	}

	s.logMsg("waiting for command exit event")

	<-c.exit

	return &taskapi.WaitResponse{
		ExitStatus: uint32(c.exitStatus),
		ExitedAt:   c.exitedAt,
	}, nil
}

func (s *Service) Stats(ctx context.Context, r *taskapi.StatsRequest) (*taskapi.StatsResponse, error) {
	return &taskapi.StatsResponse{}, nil
}

// Connect returns shim information such as the shim's pid
func (s *Service) Connect(ctx context.Context, r *taskapi.ConnectRequest) (*taskapi.ConnectResponse, error) {
	return &taskapi.ConnectResponse{
		ShimPid: uint32(os.Getpid()),
		TaskPid: uint32(s.pid),
	}, nil
}

func (s *Service) Shutdown(ctx context.Context, r *taskapi.ShutdownRequest) (*ptypes.Empty, error) {
	s.logMsg("shutdown")
	os.Exit(0)
	return empty, nil
}

func (s *Service) Cleanup(ctx context.Context) (*taskapi.DeleteResponse, error) {
	s.logMsg(fmt.Sprintf("cleanup: unmounting rootfs"))
	if err := mount.UnmountAll(s.rootfs, 0); err != nil {
		logrus.WithError(err).Warn("failed to cleanup rootfs mount")
	}
	return &taskapi.DeleteResponse{
		ExitedAt:   time.Now(),
		ExitStatus: 128 + 9, // sigkill
	}, nil
}

func (s *Service) forward(publisher events.Publisher) {
	for e := range s.events {
		if err := publisher.Publish(s.context, getTopic(s.context, e), e); err != nil {
			log.G(s.context).WithError(err).Error("post event")
		}
	}
}

func (s *Service) processExits() {
	for e := range s.exits {
		s.logMsg(fmt.Sprintf("exit event: %+v", e))
		s.checkCommand(e)
	}
}

func (s *Service) checkCommand(e runc.Exit) {
	for _, c := range s.commands {
		if c.pid == e.Pid {
			s.logMsg(fmt.Sprintf("publishing TaskExit id=%s", c.id))
			c.setExited(e)
			s.events <- &eventstypes.TaskExit{
				ContainerID: s.id,
				ID:          c.id,
				Pid:         uint32(e.Pid),
				ExitStatus:  uint32(e.Status),
				ExitedAt:    e.Timestamp,
			}
			return
		}
	}
}

func resolveRootfsCommand(rootfs, cmd string, env []string) (string, error) {
	envPath := []string{}
	for _, e := range env {
		p := strings.Split(e, "=")
		if len(p) != 2 {
			continue
		}
		if v := strings.ToUpper(p[0]); v == "PATH" {
			envPath = strings.Split(p[1], ":")
			break
		}
	}
	for _, p := range envPath {
		cmdPath := filepath.Join(rootfs, p, cmd)
		if v, _ := os.Stat(cmdPath); v != nil {
			return cmdPath, nil
		}
	}

	return "", fmt.Errorf("command not found: %s", cmd)
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
