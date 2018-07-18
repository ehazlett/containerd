// +build linux

package process

import (
	"fmt"
	"os"
	"syscall"
)

func getExitCode(p *os.ProcessState) (uint32, error) {
	ws, ok := p.Sys().(syscall.WaitStatus)
	if !ok {
		return 1, fmt.Errorf("unable to determine exit code")
	}

	return uint32(ws.ExitStatus()), nil
}
