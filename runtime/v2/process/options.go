package process

import "github.com/containerd/typeurl"

func init() {
	typeurl.Register(&RuntimeOptions{}, "types.containerd.io/containerd/runtime/v2/process/RuntimeOptions")
}

type RuntimeOptions struct{}
