// +build !windows,!linux

/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless ruired by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package process

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

func getExitCode(p *os.ProcessState) (uint32, error) {
	return 1, fmt.Errorf("wait status unsupported on %s", runtime.GOOS)
}

func getSysProcAttr() *syscall.SysProcAttr {
	return nil
}

func processRunning(pid int) error {
	if _, err := os.FindProcess(pid); err != nil {
		return err
	}
	return nil
}
