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
