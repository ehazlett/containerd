package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/containerd/containerd/runtime/v2/process"
	"github.com/containerd/containerd/runtime/v2/shim"
	"github.com/sirupsen/logrus"
)

var (
	debugFlag            bool
	namespaceFlag        string
	socketFlag           string
	addressFlag          string
	workdirFlag          string
	containerdBinaryFlag string
)

func init() {
	flag.BoolVar(&debugFlag, "debug", false, "enable debug output in logs")
	flag.StringVar(&namespaceFlag, "namespace", "", "namespace that owns the shim")
	flag.StringVar(&socketFlag, "socket", "", "abstract socket path to serve")
	flag.StringVar(&addressFlag, "address", "", "grpc address back to main containerd")
	flag.StringVar(&workdirFlag, "workdir", "", "path used to storge large temporary data")
	// currently, the `containerd publish` utility is embedded in the daemon binary.
	// The daemon invokes `containerd-shim -containerd-binary ...` with its own os.Executable() path.
	flag.StringVar(&containerdBinaryFlag, "containerd-binary", "containerd", "path to containerd binary (used for `containerd publish`)")
	flag.Parse()
}

func main() {
	debug.SetGCPercent(40)
	go func() {
		for range time.Tick(30 * time.Second) {
			debug.FreeOSMemory()
		}
	}()

	if debugFlag {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if os.Getenv("GOMAXPROCS") == "" {
		// If GOMAXPROCS hasn't been set, we default to a value of 2 to reduce
		// the number of Go stacks present in the shim.
		runtime.GOMAXPROCS(2)
	}

	cfg := &shim.Config{
		Debug:                debugFlag,
		Namespace:            namespaceFlag,
		Socket:               socketFlag,
		Address:              addressFlag,
		WorkDir:              workdirFlag,
		ContainerdBinaryPath: containerdBinaryFlag,
	}

	svc := process.NewProcessRuntime()

	s := shim.NewShim(cfg, svc)

	if err := s.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "containerd-shim: %s\n", err)
		os.Exit(1)
	}
}
