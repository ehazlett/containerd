// +build !windows,!linux

package process

import (
	"fmt"
	"os"
	"runtime"
)

func getExitCode(p *os.ProcessState) (uint32, error) {
	return 1, fmt.Errorf("wait status unsupported on %s", runtime.GOOS)
}
