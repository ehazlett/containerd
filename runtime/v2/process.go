package v2

import (
	"context"

	eventstypes "github.com/containerd/containerd/api/events"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/runtime"
	"github.com/containerd/containerd/runtime/v2/task"
	"github.com/containerd/ttrpc"
	"github.com/pkg/errors"
)

type Process struct {
	id   string
	shim *Shim
}

func (p *Process) ID() string {
	return p.id
}

func (p *Process) Kill(ctx context.Context, signal uint32, _ bool) error {
	_, err := p.shim.task.Kill(ctx, &task.KillRequest{
		Signal: signal,
		ID:     p.id,
	})
	if err != nil {
		return errdefs.FromGRPC(err)
	}
	return nil
}

func (p *Process) State(ctx context.Context) (runtime.State, error) {
	response, err := p.shim.task.State(ctx, &task.StateRequest{
		ID: p.id,
	})
	if err != nil {
		if errors.Cause(err) != ttrpc.ErrClosed {
			return runtime.State{}, errdefs.FromGRPC(err)
		}
		return runtime.State{}, errdefs.ErrNotFound
	}
	var status runtime.Status
	switch response.Status {
	case tasktypes.StatusCreated:
		status = runtime.CreatedStatus
	case tasktypes.StatusRunning:
		status = runtime.RunningStatus
	case tasktypes.StatusStopped:
		status = runtime.StoppedStatus
	case tasktypes.StatusPaused:
		status = runtime.PausedStatus
	case tasktypes.StatusPausing:
		status = runtime.PausingStatus
	}
	return runtime.State{
		Pid:        response.Pid,
		Status:     status,
		Stdin:      response.Stdin,
		Stdout:     response.Stdout,
		Stderr:     response.Stderr,
		Terminal:   response.Terminal,
		ExitStatus: response.ExitStatus,
		ExitedAt:   response.ExitedAt,
	}, nil
}

// ResizePty changes the side of the process's PTY to the provided width and height
func (p *Process) ResizePty(ctx context.Context, size runtime.ConsoleSize) error {
	_, err := p.shim.task.ResizePty(ctx, &task.ResizePtyRequest{
		ID:     p.id,
		Width:  size.Width,
		Height: size.Height,
	})
	if err != nil {
		return errdefs.FromGRPC(err)
	}
	return nil
}

// CloseIO closes the provided IO pipe for the process
func (p *Process) CloseIO(ctx context.Context) error {
	_, err := p.shim.task.CloseIO(ctx, &task.CloseIORequest{
		ID:    p.id,
		Stdin: true,
	})
	if err != nil {
		return errdefs.FromGRPC(err)
	}
	return nil
}

// Start the process
func (p *Process) Start(ctx context.Context) error {
	response, err := p.shim.task.Start(ctx, &task.StartRequest{
		ID: p.id,
	})
	if err != nil {
		return errdefs.FromGRPC(err)
	}
	p.shim.events.Publish(ctx, runtime.TaskExecStartedEventTopic, &eventstypes.TaskExecStarted{
		ContainerID: p.shim.ID(),
		Pid:         response.Pid,
		ExecID:      p.id,
	})
	return nil
}

// Wait on the process to exit and return the exit status and timestamp
func (p *Process) Wait(ctx context.Context) (*runtime.Exit, error) {
	response, err := p.shim.task.Wait(ctx, &task.WaitRequest{
		ID: p.id,
	})
	if err != nil {
		return nil, errdefs.FromGRPC(err)
	}
	return &runtime.Exit{
		Timestamp: response.ExitedAt,
		Status:    response.ExitStatus,
	}, nil
}

func (p *Process) Delete(ctx context.Context) (*runtime.Exit, error) {
	response, err := p.shim.task.Delete(ctx, &task.DeleteRequest{
		ID: p.id,
	})
	if err != nil {
		return nil, errdefs.FromGRPC(err)
	}
	return &runtime.Exit{
		Status:    response.ExitStatus,
		Timestamp: response.ExitedAt,
		Pid:       response.Pid,
	}, nil
}
