// +build !windows

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

package shim

import (
	"bytes"
	"context"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"

	eventstypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/namespaces"
	rt "github.com/containerd/containerd/runtime"
	"github.com/containerd/containerd/runtime/shim"
	shimapi "github.com/containerd/containerd/runtime/v2/task"
	"github.com/containerd/ttrpc"
	"github.com/containerd/typeurl"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type ShimClient struct {
	config  *Config
	service shimapi.TaskService
	context context.Context
	events  chan interface{}
}

func NewShimClient(config *Config, svc shimapi.TaskService) *ShimClient {
	ctx := namespaces.WithNamespace(context.Background(), config.Namespace)
	ctx = log.WithLogger(ctx, logrus.WithFields(logrus.Fields{
		"namespace": config.Namespace,
		"socket":    config.Socket,
		"pid":       os.Getpid(),
	}))
	publisher := &remoteEventsPublisher{
		address:              config.Address,
		containerdBinaryPath: config.ContainerdBinaryPath,
	}
	s := &ShimClient{
		config:  config,
		service: svc,
		context: ctx,
		events:  make(chan interface{}, 128),
	}
	go s.forward(publisher)
	return s
}

func (s *ShimClient) Serve() error {
	// start handling signals as soon as possible so that things are properly reaped
	// or if runtime exits before we hit the handler
	signals, err := setupSignals()
	if err != nil {
		return err
	}
	dump := make(chan os.Signal, 32)
	signal.Notify(dump, syscall.SIGUSR1)

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	server, err := newServer()
	if err != nil {
		return errors.Wrap(err, "failed creating server")
	}

	logrus.Debug("registering ttrpc server")
	shimapi.RegisterTaskService(server, s.service)

	if err := serve(server, s.config.Socket); err != nil {
		return err
	}
	logger := logrus.WithFields(logrus.Fields{
		"pid":       os.Getpid(),
		"path":      path,
		"namespace": s.config.Namespace,
	})
	go func() {
		for range dump {
			dumpStacks(logger)
		}
	}()
	return handleSignals(logger, signals, server, s.service)
}

func (s *ShimClient) forward(publisher events.Publisher) {
	for e := range s.events {
		if err := publisher.Publish(s.context, getTopic(s.context, e), e); err != nil {
			log.G(s.context).WithError(err).Error("post event")
		}
	}
}

func getTopic(ctx context.Context, e interface{}) string {
	switch e.(type) {
	case *eventstypes.TaskCreate:
		return rt.TaskCreateEventTopic
	case *eventstypes.TaskStart:
		return rt.TaskStartEventTopic
	case *eventstypes.TaskOOM:
		return rt.TaskOOMEventTopic
	case *eventstypes.TaskExit:
		return rt.TaskExitEventTopic
	case *eventstypes.TaskDelete:
		return rt.TaskDeleteEventTopic
	case *eventstypes.TaskExecAdded:
		return rt.TaskExecAddedEventTopic
	case *eventstypes.TaskExecStarted:
		return rt.TaskExecStartedEventTopic
	case *eventstypes.TaskPaused:
		return rt.TaskPausedEventTopic
	case *eventstypes.TaskResumed:
		return rt.TaskResumedEventTopic
	case *eventstypes.TaskCheckpointed:
		return rt.TaskCheckpointedEventTopic
	default:
		logrus.Warnf("no topic for type %#v", e)
	}
	return rt.TaskUnknownTopic
}

// serve serves the ttrpc API over a unix socket at the provided path
// this function does not block
func serve(server *ttrpc.Server, path string) error {
	var (
		l   net.Listener
		err error
	)
	if path == "" {
		l, err = net.FileListener(os.NewFile(3, "socket"))
		path = "[inherited from parent]"
	} else {
		if len(path) > 106 {
			return errors.Errorf("%q: unix socket path too long (> 106)", path)
		}
		l, err = net.Listen("unix", "\x00"+path)
	}
	if err != nil {
		return err
	}
	logrus.WithField("socket", path).Debug("serving api on unix socket")
	go func() {
		defer l.Close()
		if err := server.Serve(l); err != nil &&
			!strings.Contains(err.Error(), "use of closed network connection") {
			logrus.WithError(err).Fatal("containerd-shim: ttrpc server failure")
		}
	}()
	return nil
}

func handleSignals(logger *logrus.Entry, signals chan os.Signal, server *ttrpc.Server, sv shimapi.TaskService) error {
	var (
		termOnce sync.Once
		done     = make(chan struct{})
	)

	for {
		select {
		case <-done:
			return nil
		case s := <-signals:
			switch s {
			case unix.SIGCHLD:
				if err := shim.Reap(); err != nil {
					logger.WithError(err).Error("reap exit status")
				}
			case unix.SIGTERM, unix.SIGINT:
				go termOnce.Do(func() {
					ctx := context.TODO()
					if err := server.Shutdown(ctx); err != nil {
						logger.WithError(err).Error("failed to shutdown server")
					}
					// Ensure our child is dead if any
					sv.Kill(ctx, &shimapi.KillRequest{
						Signal: uint32(syscall.SIGKILL),
						All:    true,
					})
					idResp, err := sv.ID(ctx, &ptypes.Empty{})
					if err != nil {
						logger.WithError(err).Error("failed to get id")
					}
					sv.Delete(context.Background(), &shimapi.DeleteRequest{
						ID: idResp.ID,
					})
					close(done)
				})
			case unix.SIGPIPE:
			}
		}
	}
}

func dumpStacks(logger *logrus.Entry) {
	var (
		buf       []byte
		stackSize int
	)
	bufferLen := 16384
	for stackSize == len(buf) {
		buf = make([]byte, bufferLen)
		stackSize = runtime.Stack(buf, true)
		bufferLen *= 2
	}
	buf = buf[:stackSize]
	logger.Infof("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", buf)
}

type remoteEventsPublisher struct {
	address              string
	containerdBinaryPath string
}

func (l *remoteEventsPublisher) Publish(ctx context.Context, topic string, event events.Event) error {
	ns, _ := namespaces.Namespace(ctx)
	encoded, err := typeurl.MarshalAny(event)
	if err != nil {
		return err
	}
	data, err := encoded.Marshal()
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, l.containerdBinaryPath, "--address", l.address, "publish", "--topic", topic, "--namespace", ns)
	cmd.Stdin = bytes.NewReader(data)
	c, err := Default.Start(cmd)
	if err != nil {
		return err
	}
	status, err := Default.Wait(cmd, c)
	if err != nil {
		return err
	}
	if status != 0 {
		return errors.New("failed to publish event")
	}
	return nil
}
