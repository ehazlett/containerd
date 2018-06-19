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

//	// ID of the runtime
//	ID() string
//	// Create creates a task with the provided id and options.
//	Create(ctx context.Context, id string, opts CreateOpts) (Task, error)
//	// Get returns a task.
//	Get(context.Context, string) (Task, error)
//	// Tasks returns all the current tasks for the runtime.
//	// Any container runs at most one task at a time.
//	Tasks(context.Context) ([]Task, error)
//	// Delete removes the task in the runtime.
//	Delete(context.Context, Task) (*Exit, error)
package process

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd/events/exchange"
	"github.com/containerd/containerd/metadata"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/plugin"
	"github.com/containerd/containerd/runtime"
	"github.com/golang/protobuf/ptypes"
)

const (
	defaultRuntime = "runc"
	defaultShim    = "containerd-shim"
)

var (
	pluginID = fmt.Sprintf("%s.%s", plugin.RuntimePlugin, "process")
	empty    = &ptypes.Empty{}
)

// Runtime for a linux based system
type Runtime struct {
	root    string
	state   string
	address string

	monitor runtime.TaskMonitor
	tasks   *runtime.TaskList
	db      *metadata.DB
	events  *exchange.Exchange

	config *Config
}

// Config options for the runtime
type Config struct {
	// Shim is a path or name of binary implementing the Shim GRPC API
	Shim string `toml:"shim"`
	// Runtime is a path or name of an OCI runtime used by the shim
	Runtime string `toml:"runtime"`
}

func init() {
	plugin.Register(&plugin.Registration{
		Type:   plugin.RuntimePlugin,
		ID:     "process",
		InitFn: New,
		Requires: []plugin.Type{
			plugin.TaskMonitorPlugin,
			plugin.MetadataPlugin,
		},
		Config: &Config{
			Shim:    defaultShim,
			Runtime: defaultRuntime,
		},
	})
}

var _ = (runtime.PlatformRuntime)(&Runtime{})

// New returns a configured runtime
func New(ic *plugin.InitContext) (interface{}, error) {
	ic.Meta.Platforms = []ocispec.Platform{platforms.DefaultSpec()}

	if err := os.MkdirAll(ic.Root, 0711); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(ic.State, 0711); err != nil {
		return nil, err
	}
	monitor, err := ic.Get(plugin.TaskMonitorPlugin)
	if err != nil {
		return nil, err
	}
	m, err := ic.Get(plugin.MetadataPlugin)
	if err != nil {
		return nil, err
	}
	cfg := ic.Config.(*Config)
	r := &Runtime{
		root:    ic.Root,
		state:   ic.State,
		monitor: monitor.(runtime.TaskMonitor),
		tasks:   runtime.NewTaskList(),
		db:      m.(*metadata.DB),
		address: ic.Address,
		events:  ic.Events,
		config:  cfg,
	}
	tasks, err := r.restoreTasks(ic.Context)
	if err != nil {
		return nil, err
	}

	// TODO: need to add the tasks to the monitor
	for _, t := range tasks {
		if err := r.tasks.AddWithNamespace(t.namespace, t); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// ID of the runtime
func (r *Runtime) ID() string {
	return pluginID
}

// Create a new task
func (r *Runtime) Create(ctx context.Context, id string, opts runtime.CreateOpts) (_ runtime.Task, err error) {
	return nil, nil
}

// Get a specific task by task id
func (r *Runtime) Get(ctx context.Context, id string) (runtime.Task, error) {
	return runtime.Task{}, nil
}

// Tasks returns all tasks known to the runtime
func (r *Runtime) Tasks(ctx context.Context) ([]runtime.Task, error) {
	return nil, nil
}

// Delete a task removing all on disk state
func (r *Runtime) Delete(ctx context.Context, c runtime.Task) (*runtime.Exit, error) {
	return nil, nil
}
