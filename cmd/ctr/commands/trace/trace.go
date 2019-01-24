/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package trace

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"time"
	"unsafe"

	api "github.com/containerd/containerd/api/services/trace/v1"
	"github.com/containerd/containerd/cmd/ctr/commands"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

import "C"

// Command is a cli command to manage traces
var Command = cli.Command{
	Name:  "trace",
	Usage: "start trace",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "source",
			Usage: "path to source file",
		},
		cli.StringFlag{
			Name:  "probe-name",
			Usage: "name of probe function",
		},
		cli.StringFlag{
			Name:  "return-function-name",
			Usage: "name of return function (optional)",
			Value: "",
		},
		cli.StringFlag{
			Name:  "table-id",
			Usage: "table id",
		},
	},
	Subcommands: []cli.Command{
		kProbeCommand,
		uProbeCommand,
	},
}

var kProbeCommand = cli.Command{
	Name:  "kprobe",
	Usage: "start kprobe",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "function-name",
			Usage: "name of kernel function to trace",
		},
	},
	Action: func(context *cli.Context) error {
		id := time.Now().String()
		srcData, err := ioutil.ReadFile(context.GlobalString("source"))
		if err != nil {
			return err
		}

		client, ctx, cancel, err := commands.NewClient(context)
		if err != nil {
			return err
		}
		defer cancel()

		traceService := client.TraceService()
		functionName := context.String("function-name")
		probeRequest := &api.ProbeRequest{
			ID:     id,
			Source: string(srcData),
			ProbeConfig: &api.ProbeRequest_KprobeConfig{
				KprobeConfig: &api.KProbeConfig{
					FunctionName: functionName,
				},
			},
			ProbeName:          context.GlobalString("probe-name"),
			ReturnFunctionName: context.GlobalString("return-function-name"),
			TableID:            context.GlobalString("table-id"),
		}

		stream, err := traceService.Probe(ctx, probeRequest)
		if err != nil {
			return err
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, os.Kill)

		go func() {
			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					logrus.Fatal(err)
				}

				// TODO: this is just an example.  a "real" client would
				// do this processing for a specific trace; here we just
				// print raw data
				evt, err := processKprobeData(resp.Data)
				if err != nil {
					logrus.Fatal(err)
				}

				fmt.Printf("uid %d gid %d pid %d called fchownat on %s (return value: %d)\n", evt.Uid, evt.Gid, evt.Pid, evt.Filename, evt.ReturnValue)
			}
		}()

		<-sig

		return nil
	},
}

var uProbeCommand = cli.Command{
	Name:  "uprobe",
	Usage: "start uprobe",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Usage: "name of binary or lib to trace",
		},
		cli.StringFlag{
			Name:  "symbol",
			Usage: "name of the symbol to attach",
		},
	},
	Action: func(context *cli.Context) error {
		id := time.Now().String()
		srcData, err := ioutil.ReadFile(context.GlobalString("source"))
		if err != nil {
			return err
		}

		client, ctx, cancel, err := commands.NewClient(context)
		if err != nil {
			return err
		}
		defer cancel()

		traceService := client.TraceService()
		name := context.String("name")
		symbol := context.String("symbol")
		probeRequest := &api.ProbeRequest{
			ID:     id,
			Source: string(srcData),
			ProbeConfig: &api.ProbeRequest_UprobeConfig{
				UprobeConfig: &api.UProbeConfig{
					Name:   name,
					Symbol: symbol,
				},
			},
			ProbeName:          context.GlobalString("probe-name"),
			ReturnFunctionName: context.GlobalString("return-function-name"),
			TableID:            context.GlobalString("table-id"),
		}

		stream, err := traceService.Probe(ctx, probeRequest)
		if err != nil {
			return err
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, os.Kill)

		go func() {
			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					logrus.Fatal(err)
				}

				evt, err := processUprobeData(resp.Data)
				if err != nil {
					logrus.Fatal(err)
				}

				fmt.Printf("%+v\n", evt)
			}
		}()

		<-sig

		return nil
	},
}

// EXAMPLE -- to be removed
type chownEvent struct {
	Pid         uint32
	Uid         uint32
	Gid         uint32
	ReturnValue int32
	Filename    [256]byte
}

type chownEvt struct {
	Pid         uint32
	Uid         uint32
	Gid         uint32
	ReturnValue int32
	Filename    string
}

func processKprobeData(data []byte) (*chownEvt, error) {
	var e chownEvent
	if err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &e); err != nil {
		return nil, err
	}
	filename := (*C.char)(unsafe.Pointer(&e.Filename))
	return &chownEvt{
		Pid:         e.Pid,
		Uid:         e.Uid,
		Gid:         e.Gid,
		ReturnValue: e.ReturnValue,
		Filename:    C.GoString(filename),
	}, nil
}

type readlineEvent struct {
	Pid uint32
	Str [80]byte
}

type readlineEvt struct {
	Pid uint32
	Cmd string
}

func processUprobeData(data []byte) (*readlineEvt, error) {
	var e readlineEvent
	if err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &e); err != nil {
		return nil, err
	}
	cmd := string(e.Str[:bytes.IndexByte(e.Str[:], 0)])
	return &readlineEvt{
		Pid: e.Pid,
		Cmd: cmd,
	}, nil
}

// END EXAMPLE
