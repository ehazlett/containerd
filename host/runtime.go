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

package host

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	eventstypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/events/exchange"
	"github.com/containerd/containerd/identifiers"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/metadata"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/plugin"
	"github.com/containerd/containerd/runtime"
	ptypes "github.com/gogo/protobuf/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

var (
	pluginID = fmt.Sprintf("%s.%s", plugin.RuntimePlugin, "host")
	empty    = &ptypes.Empty{}
)

func init() {
	plugin.Register(&plugin.Registration{
		Type:   plugin.RuntimePlugin,
		ID:     "host",
		InitFn: New,
		Requires: []plugin.Type{
			plugin.TaskMonitorPlugin,
			plugin.MetadataPlugin,
		},
		Config: &Config{},
	})
}

var _ = (runtime.Runtime)(&Runtime{})

// Config options for the runtime
type Config struct {
	// RuntimeRoot is the path that shall be used by the OCI runtime for its data
	RuntimeRoot string `toml:"runtime_root"`
}

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

// ID of the runtime
func (r *Runtime) ID() string {
	return pluginID
}

// Create a new task
func (r *Runtime) Create(ctx context.Context, id string, opts runtime.CreateOpts) (_ runtime.Task, err error) {
	namespace, err := namespaces.NamespaceRequired(ctx)
	if err != nil {
		return nil, err
	}

	if err := identifiers.Validate(id); err != nil {
		return nil, errors.Wrapf(err, "invalid task id")
	}

	t := &Task{
		id:        id,
		namespace: namespace,
	}
	if err := r.tasks.Add(ctx, t); err != nil {
		return nil, err
	}
	r.events.Publish(ctx, runtime.TaskCreateEventTopic, &eventstypes.TaskCreate{
		ContainerID: id,
		Pid:         uint32(t.pid),
	})

	return t, nil
}

// Delete a task removing all on disk state
func (r *Runtime) Delete(ctx context.Context, c runtime.Task) (*runtime.Exit, error) {
	//namespace, err := namespaces.NamespaceRequired(ctx)
	//if err != nil {
	//	return nil, err
	//}
	lc, ok := c.(*Task)
	if !ok {
		return nil, fmt.Errorf("task cannot be cast as *host.Task")
	}
	if err := r.monitor.Stop(lc); err != nil {
		return nil, err
	}

	r.tasks.Delete(ctx, lc.id)
	r.events.Publish(ctx, runtime.TaskDeleteEventTopic, &eventstypes.TaskDelete{
		ContainerID: lc.id,
	})
	return &runtime.Exit{}, nil
}

// Tasks returns all tasks known to the runtime
func (r *Runtime) Tasks(ctx context.Context) ([]runtime.Task, error) {
	return r.tasks.GetAll(ctx)
}

func (r *Runtime) restoreTasks(ctx context.Context) ([]*Task, error) {
	dir, err := ioutil.ReadDir(r.state)
	if err != nil {
		return nil, err
	}
	var o []*Task
	for _, namespace := range dir {
		if !namespace.IsDir() {
			continue
		}
		name := namespace.Name()
		log.G(ctx).WithField("namespace", name).Debug("loading tasks in namespace")
		tasks, err := r.loadTasks(ctx, name)
		if err != nil {
			return nil, err
		}
		o = append(o, tasks...)
	}
	return o, nil
}

// Get a specific task by task id
func (r *Runtime) Get(ctx context.Context, id string) (runtime.Task, error) {
	return r.tasks.Get(ctx, id)
}

func (r *Runtime) loadTasks(ctx context.Context, ns string) ([]*Task, error) {
	dir, err := ioutil.ReadDir(filepath.Join(r.state, ns))
	if err != nil {
		return nil, err
	}
	var o []*Task
	for _, path := range dir {
		if !path.IsDir() {
			continue
		}
		id := path.Name()
		//bundle := loadBundle(
		//	id,
		//	filepath.Join(r.state, ns, id),
		//	filepath.Join(r.root, ns, id),
		//)
		//ctx = namespaces.WithNamespace(ctx, ns)
		//pid, _ := runc.ReadPidFile(filepath.Join(bundle.path, proc.InitPidFile))
		//s, err := bundle.NewShimClient(ctx, ns, ShimConnect(r.config, func() {
		//	err := r.cleanupAfterDeadShim(ctx, bundle, ns, id, pid)
		//	if err != nil {
		//		log.G(ctx).WithError(err).WithField("bundle", bundle.path).
		//			Error("cleaning up after dead shim")
		//	}
		//}), nil)
		//if err != nil {
		//	log.G(ctx).WithError(err).WithFields(logrus.Fields{
		//		"id":        id,
		//		"namespace": ns,
		//	}).Error("connecting to shim")
		//	err := r.cleanupAfterDeadShim(ctx, bundle, ns, id, pid)
		//	if err != nil {
		//		log.G(ctx).WithError(err).WithField("bundle", bundle.path).
		//			Error("cleaning up after dead shim")
		//	}
		//	continue
		//}
		//ropts, err := r.getRuncOptions(ctx, id)
		//if err != nil {
		//	log.G(ctx).WithError(err).WithField("id", id).
		//		Error("get runtime options")
		//	continue
		//}

		t := &Task{
			id: id,
		}

		//t, err := newTask(id, ns, pid, s, r.monitor, r.events,
		//	proc.NewRunc(ropts.RuntimeRoot, bundle.path, ns, ropts.Runtime, ropts.CriuPath, ropts.SystemdCgroup))
		//if err != nil {
		//	log.G(ctx).WithError(err).Error("loading task type")
		//	continue
		//}
		o = append(o, t)
	}
	return o, nil
}
