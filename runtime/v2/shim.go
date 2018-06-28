package v2

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	eventstypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/api/types"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/events/exchange"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/runtime"
	"github.com/containerd/containerd/runtime/v2/task"
	"github.com/containerd/containerd/sys"
	"github.com/containerd/ttrpc"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const shimBinaryFormat = "containerd-shim-%s"

// NewShim starts and returns a new shim
func NewShim(ctx context.Context, bundle *Bundle, runtime, containerdAddress string, events *exchange.Exchange) (_ *Shim, err error) {
	address, err := abstractAddress(ctx, bundle.ID)
	if err != nil {
		return nil, err
	}
	socket, err := newSocket(address)
	if err != nil {
		return nil, err
	}
	defer socket.Close()
	f, err := socket.File()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cmd, err := shimCommand(ctx, runtime, containerdAddress, bundle)
	if err != nil {
		return nil, err
	}
	cmd.ExtraFiles = append(cmd.ExtraFiles, f)

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			cmd.Process.Kill()
		}
	}()
	// make sure to wait after start
	go cmd.Wait()
	if err := writePidFile(filepath.Join(bundle.Path, "shim.pid"), cmd.Process.Pid); err != nil {
		return nil, err
	}
	log.G(ctx).WithFields(logrus.Fields{
		"pid":     cmd.Process.Pid,
		"address": address,
	}).Infof("shim %s started", cmd.Args[0])
	if err = sys.SetOOMScore(cmd.Process.Pid, sys.OOMScoreMaxKillable); err != nil {
		return nil, errors.Wrap(err, "failed to set OOM Score on shim")
	}
	// write pid file

	conn, err := connect(address, annonDialer)
	if err != nil {
		return nil, err
	}
	client := ttrpc.NewClient(conn)
	client.OnClose(conn.Close)
	return &Shim{
		bundle:  bundle,
		client:  client,
		task:    task.NewTaskClient(client),
		shimPid: cmd.Process.Pid,
		events:  events,
	}, nil
}

type Shim struct {
	bundle  *Bundle
	client  *ttrpc.Client
	task    task.TaskClient
	shimPid int
	taskPid int
	events  *exchange.Exchange
}

func (s *Shim) kill(killDuration time.Duration) error {
	dead := make(chan struct{})
	if err := unix.Kill(s.shimPid, unix.SIGTERM); err != nil && err != unix.ESRCH {
		return err
	}
	go func() {
		for {
			if err := unix.Kill(s.shimPid, 0); err == unix.ESRCH {
				close(dead)
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	select {
	case <-time.After(killDuration):
		unix.Kill(s.shimPid, unix.SIGKILL)
		<-dead
		return nil
	case <-dead:
		return nil
	}
}

// ID of the shim/task
func (s *Shim) ID() string {
	return s.bundle.ID
}

func (s *Shim) Close() error {
	return s.client.Close()
}

func (s *Shim) Delete(ctx context.Context) (*runtime.Exit, error) {
	response, err := s.task.Delete(ctx, empty)
	if err != nil {
		return nil, errdefs.FromGRPC(err)
	}
	if err := s.kill(3 * time.Second); err != nil {
		return nil, err
	}
	if err := s.bundle.Delete(); err != nil {
		return nil, err
	}
	s.events.Publish(ctx, runtime.TaskDeleteEventTopic, &eventstypes.TaskDelete{
		ContainerID: s.ID(),
		ExitStatus:  response.ExitStatus,
		ExitedAt:    response.ExitedAt,
		Pid:         response.Pid,
	})
	return &runtime.Exit{
		Status:    response.ExitStatus,
		Timestamp: response.ExitedAt,
		Pid:       response.Pid,
	}, nil
}

func (s *Shim) Create(ctx context.Context, opts runtime.CreateOpts) (Task, error) {
	request := &task.CreateTaskRequest{
		ID:         s.ID(),
		Bundle:     s.bundle.Path,
		Stdin:      opts.IO.Stdin,
		Stdout:     opts.IO.Stdout,
		Stderr:     opts.IO.Stderr,
		Terminal:   opts.IO.Terminal,
		Checkpoint: opts.Checkpoint,
		Options:    opts.Options,
	}
	for _, m := range opts.Rootfs {
		sopts.Rootfs = append(sopts.Rootfs, &types.Mount{
			Type:    m.Type,
			Source:  m.Source,
			Options: m.Options,
		})
	}
	response, err := s.task.Create(ctx, request)
	if err != nil {
		return nil, errdefs.FromGRPC(err)
	}
	s.taskPid = int(response.Pid)
	return s, nil
}

func (s *Shim) Pause(ctx context.Context) error {
	if _, err := s.task.Pause(ctx, empty); err != nil {
		return errdefs.FromGRPC(err)
	}
	s.events.Publish(ctx, runtime.TaskPausedEventTopic, &eventstypes.TaskPaused{
		ContainerID: s.ID(),
	})
	return nil
}

func (s *Shim) Resume(ctx context.Context) error {
	if _, err := s.task.Resume(ctx, empty); err != nil {
		return errdefs.FromGRPC(err)
	}
	s.events.Publish(ctx, runtime.TaskResumedEventTopic, &eventstypes.TaskResumed{
		ContainerID: s.ID(),
	})
	return nil
}

func (s *Shim) Start(ctx context.Context) error {
	response, err := s.task.Start(ctx, &task.StartRequest{
		ID: s.ID(),
	})
	if err != nil {
		return errdefs.FromGRPC(err)
	}
	s.taskPid = int(response.Pid)
	s.events.Publish(ctx, runtime.TaskStartEventTopic, &eventstypes.TaskStart{
		ContainerID: s.ID(),
		Pid:         uint32(s.taskPid),
	})
	return nil
}

func (s *Shim) Kill(ctx context.Context, signal uint32, all bool) error {
	if _, err := s.task.Kill(ctx, &task.KillRequest{
		ID:     s.ID(),
		Signal: signal,
		All:    all,
	}); err != nil {
		return errdefs.FromGRPC(err)
	}
	return nil
}

func (s *Shim) Exec(ctx context.Context, id string, opts runtime.ExecOpts) (runtime.Process, error) {

}

func (s *Shim) Pids(ctx context.Context) ([]runtime.ProcessInfo, error) {
	resp, err := s.task.Pids(ctx, &task.PidsRequest{
		ID: s.ID(),
	})
	if err != nil {
		return nil, errdefs.FromGRPC(err)
	}
	var processList []runtime.ProcessInfo
	for _, p := range resp.Processes {
		processList = append(processList, runtime.ProcessInfo{
			Pid:  p.Pid,
			Info: p.Info,
		})
	}
	return processList, nil
}

func (s *Shim) ResizePty(ctx context.Context, size runtime.ConsoleSize) error {
	_, err := s.task.ResizePty(ctx, &task.ResizePtyRequest{
		ID:     s.ID(),
		Width:  size.Width,
		Height: size.Height,
	})
	if err != nil {
		return errdefs.FromGRPC(err)
	}
	return nil
}

func (s *Shim) CloseIO(ctx context.Context) error {
	_, err := s.task.CloseIO(ctx, &task.CloseIORequest{
		ID:    s.ID(),
		Stdin: true,
	})
	if err != nil {
		return errdefs.FromGRPC(err)
	}
	return nil
}

func (s *Shim) Wait(ctx context.Context) (*runtime.Exit, error) {
	response, err := s.task.Wait(ctx, &task.WaitRequest{
		ID: s.ID(),
	})
	if err != nil {
		return nil, errdefs.FromGRPC(err)
	}
	return &runtime.Exit{
		Pid:       s.taskPid,
		Timestamp: response.ExitedAt,
		Status:    response.ExitStatus,
	}, nil
}

func shimCommand(ctx context.Context, runtime, containerdAddress string, bundle *Bundle) (*exec.Cmd, error) {
	ns, err := namespaces.NamespaceRequired(ctx)
	if err != nil {
		return nil, err
	}
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	args := []string{
		"-namespace": ns,
		"-address":   containerdAddress,
		"-containerd-binary", self,
	}
	cmd := exec.Command(getShimBinaryName(runtime), args...)
	cmd.Dir = bundle.Path
	cmd.Env = append(os.Environ(), "GOMAXPROCS=2")
	cmd.SysProcAttr = getSysProcAttr()
	return cmd, nil
}

func getShimBinaryName(runtime string) string {
	parts := strings.Split(runtime, ".")
	return fmt.Sprintf(shimBinaryFormat, parts[len(parts)-1])
}

func abstractAddress(ctx context.Context, id string) (string, error) {
	ns, err := namespaces.NamespaceRequired(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Join(string(filepath.Separator), "containerd-shim", ns, id, "shim.sock"), nil
}

func connect(address string, d func(string, time.Duration) (net.Conn, error)) (net.Conn, error) {
	return d(address, 100*time.Second)
}

func annonDialer(address string, timeout time.Duration) (net.Conn, error) {
	address = strings.TrimPrefix(address, "unix://")
	return net.DialTimeout("unix", "\x00"+address, timeout)
}

func newSocket(address string) (*net.UnixListener, error) {
	if len(address) > 106 {
		return nil, errors.Errorf("%q: unix socket path too long (> 106)", address)
	}
	l, err := net.Listen("unix", "\x00"+address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to listen to abstract unix socket %q", address)
	}
	return l.(*net.UnixListener), nil
}

func writePidFile(path string, pid int) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	tempPath := filepath.Join(filepath.Dir(path), fmt.Sprintf(".%s", filepath.Base(path)))
	f, err := os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%d", pid)
	f.Close()
	if err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
