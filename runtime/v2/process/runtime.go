package process

import (
	"context"
	"fmt"

	taskapi "github.com/containerd/containerd/runtime/v2/task"
	ptypes "github.com/gogo/protobuf/types"
)

type ProcessRuntime struct {
}

func NewProcessRuntime() *ProcessRuntime {
	return &ProcessRuntime{}
}

func (r *ProcessRuntime) State(ctx context.Context, req *taskapi.StateRequest) (*taskapi.StateResponse, error) {
	fmt.Println("STATE")
	return nil, nil
}

func (r *ProcessRuntime) Create(ctx context.Context, req *taskapi.CreateTaskRequest) (*taskapi.CreateTaskResponse, error) {
	fmt.Println("CREATE")
	return nil, nil
}

func (r *ProcessRuntime) Start(ctx context.Context, req *taskapi.StartRequest) (*taskapi.StartResponse, error) {
	fmt.Println("START")
	return nil, nil
}

func (r *ProcessRuntime) Delete(ctx context.Context, _ *ptypes.Empty) (*taskapi.DeleteResponse, error) {
	fmt.Println("DELETE")
	return nil, nil

}

func (r *ProcessRuntime) DeleteProcess(ctx context.Context, req *taskapi.DeleteProcessRequest) (*taskapi.DeleteResponse, error) {
	fmt.Println("DELETEPROCESS")
	return nil, nil

}

func (r *ProcessRuntime) Pids(ctx context.Context, req *taskapi.PidsRequest) (*taskapi.PidsResponse, error) {
	fmt.Println("PIDS")
	return nil, nil

}

func (r *ProcessRuntime) Pause(ctx context.Context, _ *ptypes.Empty) (*ptypes.Empty, error) {
	fmt.Println("PAUSE")
	return nil, nil

}

func (r *ProcessRuntime) Resume(ctx context.Context, _ *ptypes.Empty) (*ptypes.Empty, error) {
	fmt.Println("RESUME")
	return nil, nil

}

func (r *ProcessRuntime) Checkpoint(ctx context.Context, req *taskapi.CheckpointTaskRequest) (*ptypes.Empty, error) {
	fmt.Println("CHECKPOINT")
	return nil, nil

}

func (r *ProcessRuntime) Kill(ctx context.Context, req *taskapi.KillRequest) (*ptypes.Empty, error) {
	fmt.Println("KILL")
	return nil, nil

}

func (r *ProcessRuntime) Exec(ctx context.Context, req *taskapi.ExecProcessRequest) (*ptypes.Empty, error) {
	fmt.Println("EXEC")
	return nil, nil

}

func (r *ProcessRuntime) ResizePty(ctx context.Context, req *taskapi.ResizePtyRequest) (*ptypes.Empty, error) {
	fmt.Println("RESIZEPTY")
	return nil, nil

}

func (r *ProcessRuntime) CloseIO(ctx context.Context, req *taskapi.CloseIORequest) (*ptypes.Empty, error) {
	fmt.Println("CLOSEIO")
	return nil, nil

}

func (r *ProcessRuntime) Update(ctx context.Context, req *taskapi.UpdateTaskRequest) (*ptypes.Empty, error) {
	fmt.Println("UPDATE")
	return nil, nil

}

func (r *ProcessRuntime) Wait(ctx context.Context, req *taskapi.WaitRequest) (*taskapi.WaitResponse, error) {
	fmt.Println("WAIT")
	return nil, nil

}
