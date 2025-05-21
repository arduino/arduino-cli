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
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

func initInitCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	initCommand := &cobra.Command{
		Use:   "init",
		Short: i18n.Tr("Creates or updates the sketch project file."),
		Long:  i18n.Tr("Creates or updates the sketch project file."),
		Example: "" +
			"  # " + i18n.Tr("Creates or updates the sketch project file in the current directory.") + "\n" +
			"  " + os.Args[0] + " profile init\n" +
			"  " + os.Args[0] + " config init --profile Uno_profile -b arduino:avr:uno",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runInitCommand(cmd.Context(), args, srv)
		},
	}
	fqbnArg.AddToCommand(initCommand, srv)
	profileArg.AddToCommand(initCommand, srv)
	return initCommand
}

func runInitCommand(ctx context.Context, args []string, srv rpc.ArduinoCoreServiceServer) {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}

	sketchPath := arguments.InitSketchPath(path)

	inst := instance.CreateAndInit(ctx, srv)

	resp, err := srv.InitProfile(ctx, &rpc.InitProfileRequest{Instance: inst, SketchPath: sketchPath.String(), ProfileName: profileArg.Get(), Fqbn: fqbnArg.String()})
	if err != nil {
		feedback.Fatal(i18n.Tr("Error initializing the project file: %v", err), feedback.ErrGeneric)
	}
	feedback.PrintResult(profileResult{ProjectFilePath: resp.GetProjectFilePath()})
}

type profileResult struct {
	ProjectFilePath string `json:"project_path"`
}

func (ir profileResult) Data() interface{} {
	return ir
}

func (ir profileResult) String() string {
	return i18n.Tr("Project file created in: %s", ir.ProjectFilePath)
}
