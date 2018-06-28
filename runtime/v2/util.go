package v2

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/containerd/containerd/namespaces"
	"github.com/pkg/errors"
)

func shimCommand(ctx context.Context, runtime, containerdAddress string, bundle *Bundle) (*exec.Cmd, error) {
	ns, err := namespaces.NamespaceRequired(ctx)
	if err != nil {
		return nil, err
	}
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	args := []string{
		"-namespace", ns,
		"-address", containerdAddress,
		"-containerd-binary", self,
	}
	cmd := exec.Command(getShimBinaryName(runtime), args...)
	cmd.Dir = bundle.Path
	cmd.Env = append(os.Environ(), "GOMAXPROCS=2")
	cmd.SysProcAttr = getSysProcAttr()
	return cmd, nil
}

func getShimBinaryName(runtime string) string {
	parts := strings.Split(runtime, ".")
	return fmt.Sprintf(shimBinaryFormat, parts[len(parts)-1])
}

func abstractAddress(ctx context.Context, id string) (string, error) {
	ns, err := namespaces.NamespaceRequired(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Join(string(filepath.Separator), "containerd-shim", ns, id, "shim.sock"), nil
}

func connect(address string, d func(string, time.Duration) (net.Conn, error)) (net.Conn, error) {
	return d(address, 100*time.Second)
}

func annonDialer(address string, timeout time.Duration) (net.Conn, error) {
	address = strings.TrimPrefix(address, "unix://")
	return net.DialTimeout("unix", "\x00"+address, timeout)
}

func newSocket(address string) (*net.UnixListener, error) {
	if len(address) > 106 {
		return nil, errors.Errorf("%q: unix socket path too long (> 106)", address)
	}
	l, err := net.Listen("unix", "\x00"+address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to listen to abstract unix socket %q", address)
	}
	return l.(*net.UnixListener), nil
}

func writePidFile(path string, pid int) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	tempPath := filepath.Join(filepath.Dir(path), fmt.Sprintf(".%s", filepath.Base(path)))
	f, err := os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%d", pid)
	f.Close()
	if err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
