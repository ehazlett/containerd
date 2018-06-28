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

package tasks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	api "github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/api/types"
	"github.com/containerd/containerd/api/types/task"
	"github.com/containerd/containerd/archive"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/filters"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/metadata"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/plugin"
	"github.com/containerd/containerd/runtime"
	"github.com/containerd/containerd/runtime/v2"
	"github.com/containerd/containerd/services"
	"github.com/containerd/typeurl"
	ptypes "github.com/gogo/protobuf/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	_     = (api.TasksClient)(&local{})
	empty = &ptypes.Empty{}
)

func init() {
	plugin.Register(&plugin.Registration{
		Type: plugin.ServicePlugin,
		ID:   services.TasksService,
		Requires: []plugin.Type{
			plugin.RuntimePlugin,
			plugin.MetadataPlugin,
			plugin.TaskMonitorPlugin,
		},
		InitFn: initFunc,
	})
}

func initFunc(ic *plugin.InitContext) (interface{}, error) {
	rt, err := ic.GetByType(plugin.RuntimePlugin)
	if err != nil {
		return nil, err
	}

	m, err := ic.Get(plugin.MetadataPlugin)
	if err != nil {
		return nil, err
	}
	cs := m.(*metadata.DB).ContentStore()
	runtimes := make(map[string]runtime.PlatformRuntime)
	var v2Runtime *v2.TaskManager
	for _, rr := range rt {
		ri, err := rr.Instance()
		if err != nil {
			log.G(ic.Context).WithError(err).Warn("could not load runtime instance due to initialization error")
			continue
		}
		r := ri.(runtime.PlatformRuntime)
		if vt, ok := r.(*v2.TaskManager); ok {
			v2Runtime = vt
		} else {
			runtimes[r.ID()] = r
		}
	}

	if len(runtimes) == 0 {
		return nil, errors.New("no runtimes available to create task service")
	}
	monitor, err := ic.Get(plugin.TaskMonitorPlugin)
	if err != nil {
		return nil, err
	}

	return &local{
		runtimes:  runtimes,
		db:        m.(*metadata.DB),
		store:     cs,
		publisher: ic.Events,
		monitor:   monitor.(runtime.TaskMonitor),
		v2Runtime: v2Runtime,
	}, nil
}

type local struct {
	runtimes  map[string]runtime.PlatformRuntime
	db        *metadata.DB
	store     content.Store
	publisher events.Publisher

	monitor   runtime.TaskMonitor
	v2Runtime *v2.TaskManager
}

func (l *local) Create(ctx context.Context, r *api.CreateTaskRequest, _ ...grpc.CallOption) (*api.CreateTaskResponse, error) {
	var (
		checkpointPath string
		err            error
	)
	if r.Checkpoint != nil {
		checkpointPath, err = ioutil.TempDir(os.Getenv("XDG_RUNTIME_DIR"), "ctrd-checkpoint")
		if err != nil {
			return nil, err
		}
		if r.Checkpoint.MediaType != images.MediaTypeContainerd1Checkpoint {
			return nil, fmt.Errorf("unsupported checkpoint type %q", r.Checkpoint.MediaType)
		}
		reader, err := l.store.ReaderAt(ctx, ocispec.Descriptor{
			MediaType: r.Checkpoint.MediaType,
			Digest:    r.Checkpoint.Digest,
			Size:      r.Checkpoint.Size_,
		})
		if err != nil {
			return nil, err
		}
		_, err = archive.Apply(ctx, checkpointPath, content.NewReader(reader))
		reader.Close()
		if err != nil {
			return nil, err
		}
	}

	container, err := l.getContainer(ctx, r.ContainerID)
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	opts := runtime.CreateOpts{
		Spec: container.Spec,
		IO: runtime.IO{
			Stdin:    r.Stdin,
			Stdout:   r.Stdout,
			Stderr:   r.Stderr,
			Terminal: r.Terminal,
		},
		Checkpoint: checkpointPath,
		Options:    r.Options,
	}
	for _, m := range r.Rootfs {
		opts.Rootfs = append(opts.Rootfs, mount.Mount{
			Type:    m.Type,
			Source:  m.Source,
			Options: m.Options,
		})
	}
	runtime, err := l.getRuntime(container.Runtime.Name)
	if err != nil {
		return nil, err
	}
	c, err := runtime.Create(ctx, r.ContainerID, opts)
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	// TODO: fast path for getting pid on create
	if err := l.monitor.Monitor(c); err != nil {
		return nil, err
	}
	state, err := c.State(ctx)
	if err != nil {
		log.G(ctx).Error(err)
	}
	return &api.CreateTaskResponse{
		ContainerID: r.ContainerID,
		Pid:         state.Pid,
	}, nil
}

func (l *local) Start(ctx context.Context, r *api.StartRequest, _ ...grpc.CallOption) (*api.StartResponse, error) {
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	p := runtime.Process(t)
	if r.ExecID != "" {
		if p, err = t.Process(ctx, r.ExecID); err != nil {
			return nil, errdefs.ToGRPC(err)
		}
	}
	if err := p.Start(ctx); err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	state, err := p.State(ctx)
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return &api.StartResponse{
		Pid: state.Pid,
	}, nil
}

func (l *local) Delete(ctx context.Context, r *api.DeleteTaskRequest, _ ...grpc.CallOption) (*api.DeleteResponse, error) {
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	if err := l.monitor.Stop(t); err != nil {
		return nil, err
	}
	exit, err := t.Delete(ctx)
	if err != nil {
		return nil, err
	}
	return &api.DeleteResponse{
		ExitStatus: exit.Status,
		ExitedAt:   exit.Timestamp,
		Pid:        exit.Pid,
	}, nil
}

func (l *local) DeleteProcess(ctx context.Context, r *api.DeleteProcessRequest, _ ...grpc.CallOption) (*api.DeleteResponse, error) {
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	process, err := t.Process(ctx, r.ExecID)
	if err != nil {
		return nil, err
	}
	exit, err := process.Delete(ctx)
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return &api.DeleteResponse{
		ID:         r.ExecID,
		ExitStatus: exit.Status,
		ExitedAt:   exit.Timestamp,
		Pid:        exit.Pid,
	}, nil
}

func processFromContainerd(ctx context.Context, p runtime.Process) (*task.Process, error) {
	state, err := p.State(ctx)
	if err != nil {
		return nil, err
	}
	var status task.Status
	switch state.Status {
	case runtime.CreatedStatus:
		status = task.StatusCreated
	case runtime.RunningStatus:
		status = task.StatusRunning
	case runtime.StoppedStatus:
		status = task.StatusStopped
	case runtime.PausedStatus:
		status = task.StatusPaused
	case runtime.PausingStatus:
		status = task.StatusPausing
	default:
		log.G(ctx).WithField("status", state.Status).Warn("unknown status")
	}
	return &task.Process{
		ID:         p.ID(),
		Pid:        state.Pid,
		Status:     status,
		Stdin:      state.Stdin,
		Stdout:     state.Stdout,
		Stderr:     state.Stderr,
		Terminal:   state.Terminal,
		ExitStatus: state.ExitStatus,
		ExitedAt:   state.ExitedAt,
	}, nil
}

func (l *local) Get(ctx context.Context, r *api.GetRequest, _ ...grpc.CallOption) (*api.GetResponse, error) {
	task, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	p := runtime.Process(task)
	if r.ExecID != "" {
		if p, err = task.Process(ctx, r.ExecID); err != nil {
			return nil, errdefs.ToGRPC(err)
		}
	}
	t, err := processFromContainerd(ctx, p)
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return &api.GetResponse{
		Process: t,
	}, nil
}

func (l *local) List(ctx context.Context, r *api.ListTasksRequest, _ ...grpc.CallOption) (*api.ListTasksResponse, error) {
	resp := &api.ListTasksResponse{}
	for _, r := range l.runtimes {
		tasks, err := r.Tasks(ctx)
		if err != nil {
			return nil, errdefs.ToGRPC(err)
		}
		addTasks(ctx, resp, tasks)
	}
	return resp, nil
}

func addTasks(ctx context.Context, r *api.ListTasksResponse, tasks []runtime.Task) {
	for _, t := range tasks {
		tt, err := processFromContainerd(ctx, t)
		if err != nil {
			if !errdefs.IsNotFound(err) { // handle race with deletion
				log.G(ctx).WithError(err).WithField("id", t.ID()).Error("converting task to protobuf")
			}
			continue
		}
		r.Tasks = append(r.Tasks, tt)
	}
}

func (l *local) Pause(ctx context.Context, r *api.PauseTaskRequest, _ ...grpc.CallOption) (*ptypes.Empty, error) {
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	err = t.Pause(ctx)
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return empty, nil
}

func (l *local) Resume(ctx context.Context, r *api.ResumeTaskRequest, _ ...grpc.CallOption) (*ptypes.Empty, error) {
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	err = t.Resume(ctx)
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return empty, nil
}

func (l *local) Kill(ctx context.Context, r *api.KillRequest, _ ...grpc.CallOption) (*ptypes.Empty, error) {
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	p := runtime.Process(t)
	if r.ExecID != "" {
		if p, err = t.Process(ctx, r.ExecID); err != nil {
			return nil, errdefs.ToGRPC(err)
		}
	}
	if err := p.Kill(ctx, r.Signal, r.All); err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return empty, nil
}

func (l *local) ListPids(ctx context.Context, r *api.ListPidsRequest, _ ...grpc.CallOption) (*api.ListPidsResponse, error) {
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	processList, err := t.Pids(ctx)
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	var processes []*task.ProcessInfo
	for _, p := range processList {
		pInfo := task.ProcessInfo{
			Pid: p.Pid,
		}
		if p.Info != nil {
			a, err := typeurl.MarshalAny(p.Info)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to marshal process %d info", p.Pid)
			}
			pInfo.Info = a
		}
		processes = append(processes, &pInfo)
	}
	return &api.ListPidsResponse{
		Processes: processes,
	}, nil
}

func (l *local) Exec(ctx context.Context, r *api.ExecProcessRequest, _ ...grpc.CallOption) (*ptypes.Empty, error) {
	if r.ExecID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "exec id cannot be empty")
	}
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	if _, err := t.Exec(ctx, r.ExecID, runtime.ExecOpts{
		Spec: r.Spec,
		IO: runtime.IO{
			Stdin:    r.Stdin,
			Stdout:   r.Stdout,
			Stderr:   r.Stderr,
			Terminal: r.Terminal,
		},
	}); err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return empty, nil
}

func (l *local) ResizePty(ctx context.Context, r *api.ResizePtyRequest, _ ...grpc.CallOption) (*ptypes.Empty, error) {
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	p := runtime.Process(t)
	if r.ExecID != "" {
		if p, err = t.Process(ctx, r.ExecID); err != nil {
			return nil, errdefs.ToGRPC(err)
		}
	}
	if err := p.ResizePty(ctx, runtime.ConsoleSize{
		Width:  r.Width,
		Height: r.Height,
	}); err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return empty, nil
}

func (l *local) CloseIO(ctx context.Context, r *api.CloseIORequest, _ ...grpc.CallOption) (*ptypes.Empty, error) {
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	p := runtime.Process(t)
	if r.ExecID != "" {
		if p, err = t.Process(ctx, r.ExecID); err != nil {
			return nil, errdefs.ToGRPC(err)
		}
	}
	if r.Stdin {
		if err := p.CloseIO(ctx); err != nil {
			return nil, err
		}
	}
	return empty, nil
}

func (l *local) Checkpoint(ctx context.Context, r *api.CheckpointTaskRequest, _ ...grpc.CallOption) (*api.CheckpointTaskResponse, error) {
	container, err := l.getContainer(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	t, err := l.getTaskFromContainer(ctx, container)
	if err != nil {
		return nil, err
	}
	image, err := ioutil.TempDir(os.Getenv("XDG_RUNTIME_DIR"), "ctd-checkpoint")
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	defer os.RemoveAll(image)
	if err := t.Checkpoint(ctx, image, r.Options); err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	// write checkpoint to the content store
	tar := archive.Diff(ctx, "", image)
	cp, err := l.writeContent(ctx, images.MediaTypeContainerd1Checkpoint, image, tar)
	// close tar first after write
	if err := tar.Close(); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	// write the config to the content store
	data, err := container.Spec.Marshal()
	if err != nil {
		return nil, err
	}
	spec := bytes.NewReader(data)
	specD, err := l.writeContent(ctx, images.MediaTypeContainerd1CheckpointConfig, filepath.Join(image, "spec"), spec)
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return &api.CheckpointTaskResponse{
		Descriptors: []*types.Descriptor{
			cp,
			specD,
		},
	}, nil
}

func (l *local) Update(ctx context.Context, r *api.UpdateTaskRequest, _ ...grpc.CallOption) (*ptypes.Empty, error) {
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	if err := t.Update(ctx, r.Resources); err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return empty, nil
}

func (l *local) Metrics(ctx context.Context, r *api.MetricsRequest, _ ...grpc.CallOption) (*api.MetricsResponse, error) {
	panic("fix metrics filters by getting containers then tasks")
	filter, err := filters.ParseAll(r.Filters...)
	if err != nil {
		return nil, err
	}
	var resp api.MetricsResponse
	for _, r := range l.runtimes {
		tasks, err := r.Tasks(ctx)
		if err != nil {
			return nil, err
		}
		getTasksMetrics(ctx, filter, tasks, &resp)
	}
	return &resp, nil
}

func (l *local) Wait(ctx context.Context, r *api.WaitRequest, _ ...grpc.CallOption) (*api.WaitResponse, error) {
	t, err := l.getTask(ctx, r.ContainerID)
	if err != nil {
		return nil, err
	}
	p := runtime.Process(t)
	if r.ExecID != "" {
		if p, err = t.Process(ctx, r.ExecID); err != nil {
			return nil, errdefs.ToGRPC(err)
		}
	}
	exit, err := p.Wait(ctx)
	if err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return &api.WaitResponse{
		ExitStatus: exit.Status,
		ExitedAt:   exit.Timestamp,
	}, nil
}

func getTasksMetrics(ctx context.Context, filter filters.Filter, tasks []runtime.Task, r *api.MetricsResponse) {
	for _, tk := range tasks {
		if !filter.Match(filters.AdapterFunc(func(fieldpath []string) (string, bool) {
			t := tk
			switch fieldpath[0] {
			case "id":
				return t.ID(), true
			case "namespace":
			//	return t.Info().Namespace, true
			case "runtime":
				//	return t.Info().Runtime, true
			}
			return "", false
		})) {
			continue
		}
		collected := time.Now()
		stats, err := tk.Stats(ctx)
		if err != nil {
			if !errdefs.IsNotFound(err) {
				log.G(ctx).WithError(err).Errorf("collecting metrics for %s", tk.ID())
			}
			continue
		}
		r.Metrics = append(r.Metrics, &types.Metric{
			Timestamp: collected,
			ID:        tk.ID(),
			Data:      stats,
		})
	}
}

func (l *local) writeContent(ctx context.Context, mediaType, ref string, r io.Reader) (*types.Descriptor, error) {
	writer, err := l.store.Writer(ctx, content.WithRef(ref), content.WithDescriptor(ocispec.Descriptor{MediaType: mediaType}))
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	size, err := io.Copy(writer, r)
	if err != nil {
		return nil, err
	}
	if err := writer.Commit(ctx, 0, ""); err != nil {
		return nil, err
	}
	return &types.Descriptor{
		MediaType: mediaType,
		Digest:    writer.Digest(),
		Size_:     size,
	}, nil
}

func (l *local) getContainer(ctx context.Context, id string) (*containers.Container, error) {
	var container containers.Container
	if err := l.db.View(func(tx *bolt.Tx) error {
		store := metadata.NewContainerStore(tx)
		var err error
		container, err = store.Get(ctx, id)
		return err
	}); err != nil {
		return nil, errdefs.ToGRPC(err)
	}
	return &container, nil
}

func (l *local) getTask(ctx context.Context, id string) (runtime.Task, error) {
	container, err := l.getContainer(ctx, id)
	if err != nil {
		return nil, err
	}
	return l.getTaskFromContainer(ctx, container)
}

func (l *local) getTaskFromContainer(ctx context.Context, container *containers.Container) (runtime.Task, error) {
	runtime, err := l.getRuntime(container.Runtime.Name)
	if err != nil {
		return nil, errdefs.ToGRPCf(err, "runtime for task %s", container.Runtime.Name)
	}
	t, err := runtime.Get(ctx, container.ID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task %v not found", container.ID)
	}
	return t, nil
}

func (l *local) getRuntime(name string) (runtime.PlatformRuntime, error) {
	runtime, ok := l.runtimes[name]
	if !ok {
		// one runtime to rule them all
		return l.v2Runtime, nil
	}
	return runtime, nil
}
