// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package arguments

import (
	"context"

	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

// Programmer contains the programmer flag data.
// This is useful so all flags used by commands that need
// this information are consistent with each other.
type Programmer struct {
	programmer string
}

// AddToCommand adds the flags used to set the programmer to the specified Command
func (p *Programmer) AddToCommand(cmd *cobra.Command, srv rpc.ArduinoCoreServiceServer) {
	cmd.Flags().StringVarP(&p.programmer, "programmer", "P", "", tr("Programmer to use, e.g: atmel_ice"))
	cmd.RegisterFlagCompletionFunc("programmer", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return GetInstalledProgrammers(srv), cobra.ShellCompDirectiveDefault
	})
}

// String returns the programmer specified by the user, or the default programmer
// for the given board if defined.
func (p *Programmer) String(inst *rpc.Instance, fqbn string) string {
	if p.programmer != "" {
		return p.programmer
	}
	if inst == nil || fqbn == "" {
		return ""
	}
	details, err := commands.BoardDetails(context.Background(), &rpc.BoardDetailsRequest{
		Instance: inst,
		Fqbn:     fqbn,
	})
	if err != nil {
		return ""
	}
	return details.GetDefaultProgrammerId()
}

// GetProgrammer returns the programmer specified by the user
func (p *Programmer) GetProgrammer() string {
	return p.programmer
}
