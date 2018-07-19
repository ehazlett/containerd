// +build windows

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
	"github.com/Microsoft/go-winio"

	"context"
	"io"
	"sync"
	"time"
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
		dialTimeout := 3 * time.Second
		conn, err := winio.DialPipe(c.stdin, &dialTimeout)
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			io.Copy(pio.Stdin(), conn)
			pio.Stdin().Close()
			conn.Close()
			wg.Done()
		}()
	}
	if c.stdout != "" {
		c.Stdout = pio.out.w
		dialTimeout := 3 * time.Second
		conn, err := winio.DialPipe(c.stdout, &dialTimeout)
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			io.Copy(conn, pio.out.r)
			conn.Close()
			pio.out.r.Close()

			wg.Done()
		}()
	}

	cwg.Done()
	wg.Wait()
	return nil
}
