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

func initDumpCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	dumpCommand := &cobra.Command{
		Use:   "dump",
		Short: i18n.Tr("Dumps the project file."),
		Long:  i18n.Tr("Dumps the project file."),
		Example: "" +
			"  " + os.Args[0] + " profile dump\n",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runDumpCommand(cmd.Context(), args, srv)
		},
	}

	return dumpCommand
}

func runDumpCommand(ctx context.Context, args []string, srv rpc.ArduinoCoreServiceServer) {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}

	sketchPath := arguments.InitSketchPath(path)
	res := &rawResult{}
	switch feedback.GetFormat() {
	case feedback.JSON, feedback.MinifiedJSON:
		resp, err := srv.ProfileDump(ctx, &rpc.ProfileDumpRequest{SketchPath: sketchPath.String(), DumpFormat: "json"})
		if err != nil {
			feedback.Fatal(i18n.Tr("Error dumping the profile: %v", err), feedback.ErrBadArgument)
		}
		res.rawJSON = []byte(resp.GetEncodedProfile())
	case feedback.Text:
		resp, err := srv.ProfileDump(ctx, &rpc.ProfileDumpRequest{SketchPath: sketchPath.String(), DumpFormat: "yaml"})
		if err != nil {
			feedback.Fatal(i18n.Tr("Error dumping the profile: %v", err), feedback.ErrBadArgument)
		}
		res.rawYAML = []byte(resp.GetEncodedProfile())
	default:
		feedback.Fatal(i18n.Tr("Unsupported format: %s", feedback.GetFormat()), feedback.ErrBadArgument)
	}
	feedback.PrintResult(dumpResult{Config: res})
}

type rawResult struct {
	rawJSON []byte
	rawYAML []byte
}

func (r *rawResult) MarshalJSON() ([]byte, error) {
	// it is already encoded in rawJSON field
	return r.rawJSON, nil
}

type dumpResult struct {
	Config *rawResult `json:"project"`
}

func (dr dumpResult) Data() interface{} {
	return dr
}

func (dr dumpResult) String() string {
	// In case of text output do not wrap the output in outer JSON or YAML structure
	return string(dr.Config.rawYAML)
}
