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

package packages

import (
	"github.com/containerd/containerd/cmd/ctr/commands"
	"github.com/urfave/cli"
)

// Command is the cli command for managing packages
var Command = cli.Command{
	Name:    "packages",
	Aliases: []string{"package"},
	Usage:   "manage packages",
	Subcommands: cli.Commands{
		installCommand,
	},
}

var installCommand = cli.Command{
	Name:        "install",
	Usage:       "install a new package",
	ArgsUsage:   "<ref>",
	Description: "install a new package",
	Action: func(context *cli.Context) error {
		client, ctx, cancel, err := commands.NewClient(context)
		if err != nil {
			return err
		}
		defer cancel()
		ref := context.Args().First()
		return client.InstallPackage(ctx, ref)
	},
}
