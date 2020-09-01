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

package archive

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var includeBuildDir bool

// NewCommand creates a new `archive` command
func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "archive <sketchPath> <archivePath>",
		Short: "Creates a zip file containing all sketch files.",
		Long:  "Creates a zip file containing all sketch files.",
		Example: "" +
			"  " + os.Args[0] + " archive\n" +
			"  " + os.Args[0] + " archive .\n" +
			"  " + os.Args[0] + " archive . MySketchArchive.zip\n" +
			"  " + os.Args[0] + " archive /home/user/Arduino/MySketch\n" +
			"  " + os.Args[0] + " archive /home/user/Arduino/MySketch /home/user/MySketchArchive.zip",
		Args: cobra.MaximumNArgs(2),
		Run:  runArchiveCommand,
	}

	command.Flags().BoolVar(&includeBuildDir, "include-build-dir", false, "Includes build directory in the archive.")

	return command
}

func runArchiveCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino archive`")

	sketchPath := ""
	if len(args) >= 1 {
		sketchPath = args[0]
	}

	archivePath := ""
	if len(args) == 2 {
		archivePath = args[1]
	}

	_, err := commands.ArchiveSketch(context.Background(),
		&rpc.ArchiveSketchReq{
			SketchPath:      sketchPath,
			ArchivePath:     archivePath,
			IncludeBuildDir: includeBuildDir,
		})

	if err != nil {
		feedback.Errorf("Error archiving: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
}
