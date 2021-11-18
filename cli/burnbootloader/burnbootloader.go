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
	"os"

	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

var (
	fqbn       arguments.Fqbn
	port       arguments.Port
	verbose    bool
	verify     bool
	programmer string
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
		Run:     run,
	}

	fqbn.AddToCommand(burnBootloaderCommand)
	port.AddToCommand(burnBootloaderCommand)
	burnBootloaderCommand.Flags().BoolVarP(&verify, "verify", "t", false, tr("Verify uploaded binary after the upload."))
	burnBootloaderCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, tr("Turns on verbose mode."))
	burnBootloaderCommand.Flags().StringVarP(&programmer, "programmer", "P", "", tr("Use the specified programmer to upload."))
	burnBootloaderCommand.RegisterFlagCompletionFunc("programmer", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return arguments.GetInstalledProgrammers(), cobra.ShellCompDirectiveDefault
	})
	burnBootloaderCommand.Flags().BoolVar(&dryRun, "dry-run", false, tr("Do not perform the actual upload, just log out actions"))
	burnBootloaderCommand.Flags().MarkHidden("dry-run")

	return burnBootloaderCommand
}

func run(command *cobra.Command, args []string) {
	instance := instance.CreateAndInit()

	// We don't need a Sketch to upload a board's bootloader
	discoveryPort, err := port.GetPort(instance, nil)
	if err != nil {
		feedback.Errorf(tr("Error during Upload: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}

	if _, err := upload.BurnBootloader(context.Background(), &rpc.BurnBootloaderRequest{
		Instance:   instance,
		Fqbn:       fqbn.String(),
		Port:       discoveryPort.ToRPC(),
		Verbose:    verbose,
		Verify:     verify,
		Programmer: programmer,
		DryRun:     dryRun,
	}, os.Stdout, os.Stderr); err != nil {
		feedback.Errorf(tr("Error during Upload: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
	os.Exit(0)
}
