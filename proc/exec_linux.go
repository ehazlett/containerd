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

package proc

import (
	"context"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

func (e *execProcess) kill(ctx context.Context, sig uint32, _ bool) error {
	pid := e.pid
	if pid != 0 {
		if err := unix.Kill(pid, syscall.Signal(sig)); err != nil {
			return errors.Wrapf(checkKillError(err), "exec kill error")
		}
	}
	return nil
}

func (e *execProcess) Status(ctx context.Context) (string, error) {
	s, err := e.parent.Status(ctx)
	if err != nil {
		return "", err
	}
	// if the container as a whole is in the pausing/paused state, so are all
	// other processes inside the container, use container state here
	switch s {
	case "paused", "pausing":
		return s, nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	// if we don't have a pid then the exec process has just been created
	if e.pid == 0 {
		return "created", nil
	}
	// if we have a pid and it can be signaled, the process is running
	if err := unix.Kill(e.pid, 0); err == nil {
		return "running", nil
	}
	// else if we have a pid but it can nolonger be signaled, it has stopped
	return "stopped", nil
}
