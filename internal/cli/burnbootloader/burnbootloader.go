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

package burnbootloader

import (
	"context"
	"errors"
	"os"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbn       arguments.Fqbn
	port       arguments.Port
	verbose    bool
	verify     bool
	programmer arguments.Programmer
	dryRun     bool
	tr         = i18n.Tr
)

// NewCommand created a new `burn-bootloader` command
func NewCommand() *cobra.Command {
	burnBootloaderCommand := &cobra.Command{
		Use:     "burn-bootloader",
		Short:   tr("Upload the bootloader."),
		Long:    tr("Upload the bootloader on the board using an external programmer."),
		Example: "  " + os.Args[0] + " burn-bootloader -b arduino:avr:uno -P atmel_ice",
		Args:    cobra.MaximumNArgs(1),
		Run:     runBootloaderCommand,
	}

	fqbn.AddToCommand(burnBootloaderCommand)
	port.AddToCommand(burnBootloaderCommand)
	programmer.AddToCommand(burnBootloaderCommand)
	burnBootloaderCommand.Flags().BoolVarP(&verify, "verify", "t", false, tr("Verify uploaded binary after the upload."))
	burnBootloaderCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, tr("Turns on verbose mode."))
	burnBootloaderCommand.Flags().BoolVar(&dryRun, "dry-run", false, tr("Do not perform the actual upload, just log out actions"))
	burnBootloaderCommand.Flags().MarkHidden("dry-run")

	return burnBootloaderCommand
}

func runBootloaderCommand(command *cobra.Command, args []string) {
	instance := instance.CreateAndInit()

	logrus.Info("Executing `arduino-cli burn-bootloader`")

	// We don't need a Sketch to upload a board's bootloader
	discoveryPort, err := port.GetPort(instance, "", "")
	if err != nil {
		feedback.Fatal(tr("Error during Upload: %v", err), feedback.ErrGeneric)
	}

	stdOut, stdErr, res := feedback.OutputStreams()
	if _, err := upload.BurnBootloader(context.Background(), &rpc.BurnBootloaderRequest{
		Instance:   instance,
		Fqbn:       fqbn.String(),
		Port:       discoveryPort,
		Verbose:    verbose,
		Verify:     verify,
		Programmer: programmer.String(instance, fqbn.String()),
		DryRun:     dryRun,
	}, stdOut, stdErr); err != nil {
		errcode := feedback.ErrGeneric
		if errors.Is(err, &arduino.ProgrammerRequiredForUploadError{}) {
			errcode = feedback.ErrMissingProgrammer
		}
		if errors.Is(err, &arduino.MissingProgrammerError{}) {
			errcode = feedback.ErrMissingProgrammer
		}
		feedback.Fatal(tr("Error during Upload: %v", err), errcode)
	}
	feedback.PrintResult(res())
}
