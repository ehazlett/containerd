/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package process

import (
	"context"
	"fmt"

	taskapi "github.com/containerd/containerd/runtime/v2/task"
	ptypes "github.com/gogo/protobuf/types"
)

type ProcessRuntime struct {
	id string
}

func NewProcessRuntime() *ProcessRuntime {
	return &ProcessRuntime{
		id: "test",
	}
}

func (r *ProcessRuntime) ID(ctx context.Context, _ *ptypes.Empty) (*taskapi.IDResponse, error) {
	return &taskapi.IDResponse{ID: r.id}, nil
}

func (r *ProcessRuntime) State(ctx context.Context, req *taskapi.StateRequest) (*taskapi.StateResponse, error) {
	fmt.Println("STATE")
	return &taskapi.StateResponse{}, nil
}

func (r *ProcessRuntime) Create(ctx context.Context, req *taskapi.CreateTaskRequest) (*taskapi.CreateTaskResponse, error) {
	fmt.Println("CREATE")
	return &taskapi.CreateTaskResponse{}, nil
}

func (r *ProcessRuntime) Start(ctx context.Context, req *taskapi.StartRequest) (*taskapi.StartResponse, error) {
	fmt.Println("START")
	return &taskapi.StartResponse{}, nil
}

func (r *ProcessRuntime) Delete(ctx context.Context, req *taskapi.DeleteRequest) (*taskapi.DeleteResponse, error) {
	fmt.Println("DELETE")
	return &taskapi.DeleteResponse{}, nil

}

func (r *ProcessRuntime) Pids(ctx context.Context, req *taskapi.PidsRequest) (*taskapi.PidsResponse, error) {
	fmt.Println("PIDS")
	return &taskapi.PidsResponse{}, nil

}

func (r *ProcessRuntime) Pause(ctx context.Context, _ *ptypes.Empty) (*ptypes.Empty, error) {
	fmt.Println("PAUSE")
	return &ptypes.Empty{}, nil

}

func (r *ProcessRuntime) Resume(ctx context.Context, _ *ptypes.Empty) (*ptypes.Empty, error) {
	fmt.Println("RESUME")
	return &ptypes.Empty{}, nil

}

func (r *ProcessRuntime) Checkpoint(ctx context.Context, req *taskapi.CheckpointTaskRequest) (*ptypes.Empty, error) {
	fmt.Println("CHECKPOINT")
	return &ptypes.Empty{}, nil

}

func (r *ProcessRuntime) Kill(ctx context.Context, req *taskapi.KillRequest) (*ptypes.Empty, error) {
	fmt.Println("KILL")
	return &ptypes.Empty{}, nil

}

func (r *ProcessRuntime) Exec(ctx context.Context, req *taskapi.ExecProcessRequest) (*ptypes.Empty, error) {
	fmt.Println("EXEC")
	return &ptypes.Empty{}, nil

}

func (r *ProcessRuntime) ResizePty(ctx context.Context, req *taskapi.ResizePtyRequest) (*ptypes.Empty, error) {
	fmt.Println("RESIZEPTY")
	return &ptypes.Empty{}, nil

}

func (r *ProcessRuntime) CloseIO(ctx context.Context, req *taskapi.CloseIORequest) (*ptypes.Empty, error) {
	fmt.Println("CLOSEIO")
	return &ptypes.Empty{}, nil

}

func (r *ProcessRuntime) Update(ctx context.Context, req *taskapi.UpdateTaskRequest) (*ptypes.Empty, error) {
	fmt.Println("UPDATE")
	return &ptypes.Empty{}, nil

}

func (r *ProcessRuntime) Wait(ctx context.Context, req *taskapi.WaitRequest) (*taskapi.WaitResponse, error) {
	fmt.Println("WAIT")
	return &taskapi.WaitResponse{}, nil

}

func (r *ProcessRuntime) Stats(ctx context.Context, req *taskapi.StatsRequest) (*taskapi.StatsResponse, error) {
	fmt.Println("STATS")
	return &taskapi.StatsResponse{}, nil

}
