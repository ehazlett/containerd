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
	"context"
	"io"
	"sync"
	"syscall"

	"github.com/containerd/fifo"
)

func setupIO(ctx context.Context, id string, c *command, cwg *sync.WaitGroup) error {
	wg := &sync.WaitGroup{}
	pio, err := NewIO()
	if err != nil {
		return err
	}
	c.pio = pio

	if c.stdin != "" {
		c.Stdin = pio.in.r
		in, err := fifo.OpenFifo(ctx, c.stdin, syscall.O_RDONLY, 0)
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			io.Copy(pio.Stdin(), in)
			pio.Stdin().Close()
			in.Close()
			wg.Done()
		}()
	}
	if c.stdout != "" {
		c.Stdout = pio.out.w
		outw, err := fifo.OpenFifo(ctx, c.stdout, syscall.O_WRONLY, 0)
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			io.Copy(outw, pio.out.r)
			outw.Close()
			pio.out.r.Close()

			wg.Done()
		}()
	}

	cwg.Done()
	wg.Wait()
	return nil
}
