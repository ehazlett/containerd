package host

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/runtime"
	"github.com/gogo/protobuf/types"
)

type Task struct {
	id        string
	namespace string
	pid       int
}

func (t *Task) ID() string {
	return t.id
}

// State returns the process state
func (t *Task) State(context.Context) (runtime.State, error) {
	return runtime.State{}, fmt.Errorf("not implemented")
}

// Kill signals a container
func (t *Task) Kill(context.Context, uint32, bool) error {
	return fmt.Errorf("not implemented")
}

// Pty resizes the processes pty/console
func (t *Task) ResizePty(context.Context, runtime.ConsoleSize) error {
	return fmt.Errorf("not implemented")
}

// CloseStdin closes the processes stdin
func (t *Task) CloseIO(context.Context) error {
	return fmt.Errorf("not implemented")
}

// Start the container's user defined process
func (t *Task) Start(context.Context) error {
	return fmt.Errorf("not implemented")
}

// Wait for the process to exit
func (t *Task) Wait(context.Context) (*runtime.Exit, error) {
	return nil, fmt.Errorf("not implemented")
}

// Information of the container
func (t *Task) Info() runtime.TaskInfo {
	return runtime.TaskInfo{}
}

// Pause pauses the container process
func (t *Task) Pause(context.Context) error {
	return fmt.Errorf("not implemented")
}

// Resume unpauses the container process
func (t *Task) Resume(context.Context) error {

	return fmt.Errorf("not implemented")
}

// Exec adds a process into the container
func (t *Task) Exec(context.Context, string, runtime.ExecOpts) (runtime.Process, error) {

	return nil, fmt.Errorf("not implemented")
}

// Pids returns all pids
func (t *Task) Pids(context.Context) ([]runtime.ProcessInfo, error) {

	return nil, fmt.Errorf("not implemented")
}

// Checkpoint checkpoints a container to an image with live system data
func (t *Task) Checkpoint(context.Context, string, *types.Any) error {

	return fmt.Errorf("not implemented")
}

// DeleteProcess deletes a specific exec process via its id
func (t *Task) DeleteProcess(context.Context, string) (*runtime.Exit, error) {
	return nil, fmt.Errorf("not implemented")
}

// Update sets the provided resources to a running task
func (t *Task) Update(context.Context, *types.Any) error {
	return fmt.Errorf("not implemented")
}

// Process returns a process within the task for the provided id
func (t *Task) Process(context.Context, string) (runtime.Process, error) {
	return nil, fmt.Errorf("not implemented")
}

// Metrics returns runtime specific metrics for a task
func (t *Task) Metrics(context.Context) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
