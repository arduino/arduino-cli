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
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

// Fqbn contains the fqbn flag data.
// This is useful so all flags used by commands that need
// this information are consistent with each other.
type Fqbn struct {
	fqbn         string
	boardOptions []string // List of boards specific options separated by commas. Or can be used multiple times for multiple options.
}

// AddToCommand adds the flags used to set fqbn to the specified Command
func (f *Fqbn) AddToCommand(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.fqbn, "fqbn", "b", "", tr("Fully Qualified Board Name, e.g.: arduino:avr:uno"))
	cmd.RegisterFlagCompletionFunc("fqbn", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return GetInstalledBoards(), cobra.ShellCompDirectiveDefault
	})
	cmd.Flags().StringSliceVar(&f.boardOptions, "board-options", []string{},
		tr("List of board options separated by commas. Or can be used multiple times for multiple options."))
}

// String returns the fqbn with the board options if there are any
func (f *Fqbn) String() string {
	// If boardOptions are passed with the "--board-options" flags then add them along with the fqbn
	// This way it's possible to use either the legacy way (appending board options directly to the fqbn),
	// or the new and more elegant way (using "--board-options"), even using multiple "--board-options" works.
	if f.fqbn != "" && len(f.boardOptions) != 0 {
		return f.fqbn + ":" + strings.Join(f.boardOptions, ",")
	}
	return f.fqbn
}

// Set sets the fqbn
func (f *Fqbn) Set(fqbn string) {
	f.fqbn = fqbn
}

// CalculateFQBNAndPort calculate the FQBN and Port metadata based on
// parameters provided by the user.
// This determine the FQBN based on:
// - the value of the FQBN flag if explicitly specified, otherwise
// - the default FQBN value in sketch.yaml (`default_fqbn` key) if available, otherwise
// - it tries to autodetect the board connected to the given port flags
// If all above methods fails, it returns the empty string.
// The Port metadata are always returned except if:
//   - the port is not found, in this case nil is returned
//   - the FQBN autodetection fail, in this case the function prints an error and
//     terminates the execution
func CalculateFQBNAndPort(portArgs *Port, fqbnArg *Fqbn, instance *rpc.Instance, defaultFQBN, defaultAddress, defaultProtocol string) (string, *rpc.Port) {
	fqbn := fqbnArg.String()
	if fqbn == "" {
		fqbn = defaultFQBN
	}
	if fqbn == "" {
		if portArgs == nil || portArgs.address == "" {
			feedback.FatalError(&arduino.MissingFQBNError{}, feedback.ErrGeneric)
		}
		fqbn, port := portArgs.DetectFQBN(instance)
		if fqbn == "" {
			feedback.FatalError(&arduino.MissingFQBNError{}, feedback.ErrGeneric)
		}
		return fqbn, port
	}

	port, err := portArgs.GetPort(instance, defaultAddress, defaultProtocol)
	if err != nil {
		feedback.Fatal(tr("Error getting port metadata: %v", err), feedback.ErrGeneric)
	}
	return fqbn, port.ToRPC()
}
