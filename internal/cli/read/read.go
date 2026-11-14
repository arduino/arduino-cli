// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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

package read

import (
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

// NewCommand creates the read command. The platform must explicitly support
// the selected memory, format, and bootloader/programmer route.
func NewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var fqbn arguments.Fqbn
	var port arguments.Port
	var programmer arguments.Programmer
	var memory, format, output string
	var verbose, dryRun bool
	cmd := &cobra.Command{
		Use:   "read",
		Short: i18n.Tr("Read board memory to a file."),
		Long:  i18n.Tr("Read board memory using a platform-provided recipe. No sketch is required. Existing output files are never overwritten. Reading may reset or halt the board; it does not recover sketch source code."),
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := cmd.Context()
			inst := instance.CreateAndInit(ctx, srv)
			selectedPort, err := port.GetPort(ctx, inst, srv, "", "", nil)
			if err != nil {
				feedback.Fatal(i18n.Tr("Error getting port metadata: %v", err), feedback.ErrGeneric)
			}
			stdout, stderr, stdioResult := feedback.OutputStreams()
			stream, response := commands.ReadMemoryToServerStreams(ctx, stdout, stderr)
			if err := srv.ReadMemory(&rpc.ReadMemoryRequest{
				Instance: inst, Fqbn: fqbn.String(), Port: selectedPort,
				Programmer: programmer.GetProgrammer(), Memory: memory, Format: format,
				OutputFile: output, Verbose: verbose, DryRun: dryRun,
			}, stream); err != nil {
				feedback.Fatal(i18n.Tr("Error reading memory: %v", err), feedback.ErrGeneric)
			}
			ioResult := stdioResult()
			res := response()
			feedback.PrintResult(&readResult{OutputFile: res.GetOutputFile(), Size: res.GetSize(), DryRun: res.GetDryRun(), Stdout: ioResult.Stdout, Stderr: ioResult.Stderr})
		},
	}
	fqbn.AddToCommand(cmd, srv)
	port.AddToCommand(cmd, srv)
	cmd.Flags().Lookup("port").Usage = i18n.Tr("Board or programmer port address, e.g.: COM3 or /dev/ttyACM2")
	cmd.Flags().Lookup("protocol").Usage = i18n.Tr("Board or programmer port protocol, e.g.: serial")
	programmer.AddToCommand(cmd, srv)
	cmd.Flags().StringVar(&memory, "memory", "flash", i18n.Tr("Memory identifier supported by the platform, e.g. flash or eeprom."))
	cmd.Flags().StringVar(&format, "file-format", "bin", i18n.Tr("Output format supported by the platform, e.g. bin or ihex."))
	cmd.Flags().StringVarP(&output, "output", "o", "", i18n.Tr("New output file; must not already exist."))
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, i18n.Tr("Turns on verbose mode."))
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, i18n.Tr("Print the read recipe without running it or creating output files."))
	cmd.MarkFlagRequired("fqbn")
	cmd.MarkFlagRequired("output")
	return cmd
}

type readResult struct {
	OutputFile string `json:"output_file"`
	Size       int64  `json:"size"`
	DryRun     bool   `json:"dry_run"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
}

func (r *readResult) Data() any { return r }
func (r *readResult) String() string {
	if r.DryRun {
		return i18n.Tr("Readback dry run complete; no output file created.")
	}
	return i18n.Tr("Read %d bytes to %s", r.Size, r.OutputFile)
}
