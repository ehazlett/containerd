package shim

type Config struct {
	Debug                bool
	Namespace            string
	Socket               string
	Address              string
	WorkDir              string
	ContainerdBinaryPath string
}
