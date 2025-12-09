// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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

package profile

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

func initProfileSetDefaultCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var sketchDir string
	setDefaultCommand := &cobra.Command{
		Use:   "set-default",
		Short: i18n.Tr("Set the default build profile."),
		Long:  i18n.Tr("Set the default build profile."),
		Example: "" +
			"  " + os.Args[0] + " profile set-default my_profile\n",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runSetDefaultCommand(cmd.Context(), args, srv, sketchDir)
		},
	}
	setDefaultCommand.Flags().StringVar(&sketchDir, "sketch-path", "", i18n.Tr("Location of the sketch."))
	return setDefaultCommand
}

func runSetDefaultCommand(ctx context.Context, args []string, srv rpc.ArduinoCoreServiceServer, sketchDir string) {
	profileName := args[0]
	sketchPath := arguments.InitSketchPath(sketchDir)

	_, err := srv.ProfileSetDefault(ctx, &rpc.ProfileSetDefaultRequest{SketchPath: sketchPath.String(), ProfileName: profileName})
	if err != nil {
		feedback.Fatal(i18n.Tr("Cannot set %s as default profile: %v", profileName, err), feedback.ErrGeneric)
	}
	feedback.PrintResult(&profileSetDefaultResult{
		DefaultProfile: profileName,
	})
}

type profileSetDefaultResult struct {
	DefaultProfile string `json:"default_profile"`
}

func (r *profileSetDefaultResult) String() string {
	return i18n.Tr("Default profile set to: %s", r.DefaultProfile)
}

func (r *profileSetDefaultResult) Data() any {
	return r
}
