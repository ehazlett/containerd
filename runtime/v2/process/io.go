package process

import (
	"io"
	"os"
	"sync"
)

type pipeIO struct {
	in  *pipe
	out *pipe
	err *pipe
}

type pipe struct {
	r *os.File
	w *os.File
}

var (
	bufPool = sync.Pool{
		New: func() interface{} {
			buffer := make([]byte, 32<<10)
			return &buffer
		},
	}
)

func (p *pipe) Close() error {
	err := p.r.Close()
	if werr := p.w.Close(); err == nil {
		err = werr
	}
	return err
}

func newPipe() (*pipe, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	return &pipe{
		r: r,
		w: w,
	}, nil
}

func NewIO() (i *pipeIO, err error) {
	var pipes []*pipe
	defer func() {
		if err != nil {
			for _, p := range pipes {
				p.Close()
			}
		}
	}()
	stdin, err := newPipe()
	if err != nil {
		return nil, err
	}
	pipes = append(pipes, stdin)

	stdout, err := newPipe()
	if err != nil {
		return nil, err
	}
	pipes = append(pipes, stdout)

	stderr, err := newPipe()
	if err != nil {
		return nil, err
	}
	pipes = append(pipes, stderr)

	return &pipeIO{
		in:  stdin,
		out: stdout,
		err: stderr,
	}, nil
}

func (i *pipeIO) Stdin() io.WriteCloser {
	return i.in.w
}

func (i *pipeIO) Stdout() io.ReadCloser {
	return i.out.r
}

func (i *pipeIO) Stderr() io.ReadCloser {
	return i.err.r
}

func (i *pipeIO) Close() error {
	var err error
	for _, v := range []*pipe{
		i.in,
		i.out,
		i.err,
	} {
		if cerr := v.Close(); err == nil {
			err = cerr
		}
	}
	return err
}
