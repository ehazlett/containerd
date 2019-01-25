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

package trace

import (
	"context"
	"fmt"
	"sync"

	api "github.com/containerd/containerd/api/services/trace/v1"
	"github.com/containerd/containerd/plugin"
	ptypes "github.com/gogo/protobuf/types"
	bpf "github.com/iovisor/gobpf/bcc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	_ = (api.TraceServer)(&service{})
)

type module struct {
	m       *bpf.Module
	perfMap *bpf.PerfMap
	ch      chan []byte
}

func init() {
	plugin.Register(&plugin.Registration{
		Type: plugin.GRPCPlugin,
		ID:   "trace",
		InitFn: func(ic *plugin.InitContext) (interface{}, error) {
			return newService()
		},
	})
}

type service struct {
	mu     *sync.Mutex
	probes map[string]*module
}

func newService() (*service, error) {
	return &service{
		mu:     &sync.Mutex{},
		probes: make(map[string]*module),
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterTraceServer(server, s)
	return nil
}

func (s *service) Probe(req *api.ProbeRequest, srv api.Trace_ProbeServer) error {
	m := bpf.NewModule(req.Source, []string{})

	// TODO: defer func to Close module.  currently if the Load or Attach fails
	// a device busy is reported until the daemon is restarted
	probeConfig := req.GetProbeConfig()
	switch t := probeConfig.(type) {
	case *api.ProbeRequest_KprobeConfig:
		kprobe, err := m.LoadKprobe(req.ProbeName)
		if err != nil {
			return errors.Wrapf(err, "error loading kprobe for %s", req.ProbeName)
		}
		syscallName := t.KprobeConfig.Syscall
		if err := m.AttachKprobe(syscallName, kprobe); err != nil {
			return errors.Wrapf(err, "error attaching to syscall %s", syscallName)
		}
		if r := req.ReturnFunctionName; r != "" {
			returnProbe, err := m.LoadKprobe(r)
			if err != nil {
				return err
			}
			if err := m.AttachKretprobe(syscallName, returnProbe); err != nil {
				return errors.Wrapf(err, "error attaching return probe %s syscall %s", r, syscallName)
			}
		}
	case *api.ProbeRequest_UprobeConfig:
		uprobe, err := m.LoadUprobe(req.ProbeName)
		if err != nil {
			return errors.Wrapf(err, "error loading uprobe for %s", req.ProbeName)
		}
		name := t.UprobeConfig.Name
		symbol := t.UprobeConfig.Symbol
		if r := req.ReturnFunctionName; r == "" {
			// TODO: enable pid instead of all (-1)
			if err := m.AttachUprobe(name, symbol, uprobe, -1); err != nil {
				return errors.Wrapf(err, "error attaching to uprobe name=%s symbol=%s", name, symbol)
			}
		} else {
			returnProbe, err := m.LoadUprobe(r)
			if err != nil {
				return err
			}
			if err := m.AttachUretprobe(name, symbol, returnProbe, -1); err != nil {
				return errors.Wrapf(err, "error attaching return probe %s", r)
			}
		}
	default:
		return fmt.Errorf("unknown probe config: %v", t)
	}

	table := bpf.NewTable(m.TableId(req.TableID), m)
	ch := make(chan []byte)
	perfMap, err := bpf.InitPerfMap(table, ch)
	if err != nil {
		return errors.Wrap(err, "error initializing perf map")
	}

	doneCh := make(chan bool, 1)
	go func() {
		var err error
		for {
			data := <-ch
			err = srv.Send(&api.ProbeResponse{
				Data: data,
			})
			if err != nil {
				logrus.Errorf("error sending trace data to client: %s", err)
				return
			}
		}
	}()

	s.mu.Lock()
	s.probes[req.ID] = &module{
		m:       m,
		perfMap: perfMap,
		ch:      ch,
	}

	perfMap.Start()
	s.mu.Unlock()

	<-doneCh

	return nil
}

func (s *service) Unload(ctx context.Context, req *api.UnloadRequest) (*ptypes.Empty, error) {
	logrus.WithField("id", req.ID).Debug("unloading trace")
	s.unloadModule(req.ID)
	return &ptypes.Empty{}, nil
}

func (s *service) unloadModule(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if mod, ok := s.probes[id]; ok {
		mod.perfMap.Stop()
		mod.m.Close()
		close(mod.ch)
		delete(s.probes, id)
	}
	logrus.WithField("id", id).Debug("unloaded trace")
}
