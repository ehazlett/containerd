// +build linux

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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

var (
	ldConfPath = "/etc/ld.so.conf.d"
)

func getExitCode(p *os.ProcessState) (uint32, error) {
	ws, ok := p.Sys().(syscall.WaitStatus)
	if !ok {
		return 1, fmt.Errorf("unable to determine exit code")
	}

	return uint32(ws.ExitStatus()), nil
}

func getSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func processRunning(pid int) error {
	return syscall.Kill(pid, syscall.Signal(0))
}

func (s *Service) adaptCommandEnvironment(env []string) ([]string, error) {
	updated := []string{}
	ldPath := false

	for _, e := range env {
		p := strings.Split(e, "=")
		if len(p) == 1 {
			updated = append(updated, e)
			continue
		}
		k := strings.ToUpper(p[0])
		v := strings.Join(p[1:], "=")
		res := ""
		switch k {
		case "PATH":
			res = adaptRootfsPath(s.rootfs, v)
		case "LD_LIBRARY_PATH":
			// ignore adapting if present
			ldPath = true
		default:
			res = v
		}

		updated = append(updated, k+"="+res)
	}

	// check if LD_LIBRARY_PATH was present; if not inject
	if !ldPath {
		r, err := resolveRootfsLDLibraryPath(s.rootfs)
		if err != nil {
			return nil, err
		}
		// add host
		h, err := resolveRootfsLDLibraryPath("")
		if err != nil {
			return nil, err
		}
		if r != "" {
			updated = append(updated, "LD_LIBRARY_PATH="+r+h)
		}
	}

	return updated, nil
}

func adaptRootfsPath(rootfs, val string) string {
	dirs := []string{}
	p := strings.Split(val, ":")
	for _, d := range p {
		dirs = append(dirs, filepath.Join(rootfs, d))
	}

	return strings.Join(dirs, ":")
}

func resolveRootfsLDLibraryPath(rootfs string) (string, error) {
	ldPath := filepath.Join(rootfs, ldConfPath)
	if _, err := os.Stat(ldPath); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", errors.Wrapf(err, "config path not found for LD_LIBRARY_PATH: %s", ldPath)
	}
	updated := []string{}
	confs, err := ioutil.ReadDir(ldPath)
	if err != nil {
		return "", errors.Wrapf(err, "error reading LD_LIBRARY_PATH conf dir: %s", ldPath)
	}
	for _, c := range confs {
		config := filepath.Join(ldPath, c.Name())
		f, err := os.Open(config)
		if err != nil {
			return "", errors.Wrapf(err, "error reading LD_LIBRARY_PATH conf dir: %s", config)
		}
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			v := s.Text()
			// skip comments
			if strings.Index(v, "#") == 0 {
				continue
			}
			updated = append(updated, filepath.Join(rootfs, v))
		}
	}

	return strings.Join(updated, ":"), nil
}
