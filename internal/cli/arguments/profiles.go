// This file is part of arduino-cli.
//
// Copyright 2020-2022 ARDUINO SA (http://www.arduino.cc/)
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
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

// Profile contains the profile flag data.
// This is useful so all flags used by commands that need
// this information are consistent with each other.
type Profile struct {
	profile string
}

// AddToCommand adds the flags used to set fqbn to the specified Command
func (f *Profile) AddToCommand(cmd *cobra.Command, srv rpc.ArduinoCoreServiceServer) {
	cmd.Flags().StringVarP(&f.profile, "profile", "m", "", i18n.Tr("Sketch profile to use"))
	cmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var sketchProfile string
		if len(args) > 0 {
			sketchProfile = args[0]
		}
		return GetSketchProfiles(cmd.Context(), srv, sketchProfile), cobra.ShellCompDirectiveDefault
	})
}

// Get returns the profile name
func (f *Profile) Get() string {
	return f.profile
}

// String returns the profile name
func (f *Profile) String() string {
	return f.profile
}

// Set sets the profile
func (f *Profile) Set(profile string) {
	f.profile = profile
}
