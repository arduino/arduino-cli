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

package sketch

import (
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	sk "github.com/arduino/arduino-cli/commands/sketch"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var includeBuildDir bool

// initArchiveCommand creates a new `archive` command
func initArchiveCommand() *cobra.Command {
	archiveCommand := &cobra.Command{
		Use:   fmt.Sprintf("archive <%s> <%s>", tr("sketchPath"), tr("archivePath")),
		Short: tr("Creates a zip file containing all sketch files."),
		Long:  tr("Creates a zip file containing all sketch files."),
		Example: "" +
			"  " + os.Args[0] + " archive\n" +
			"  " + os.Args[0] + " archive .\n" +
			"  " + os.Args[0] + " archive . MySketchArchive.zip\n" +
			"  " + os.Args[0] + " archive /home/user/Arduino/MySketch\n" +
			"  " + os.Args[0] + " archive /home/user/Arduino/MySketch /home/user/MySketchArchive.zip",
		Args: cobra.MaximumNArgs(2),
		Run:  runArchiveCommand,
	}

	archiveCommand.Flags().BoolVar(&includeBuildDir, "include-build-dir", false, tr("Includes %s directory in the archive.", "build"))

	return archiveCommand
}

func runArchiveCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino sketch archive`")

	sketchPath := paths.New(".")
	if len(args) >= 1 {
		sketchPath = paths.New(args[0])
	}

	arguments.WarnDeprecatedFiles(sketchPath)

	archivePath := ""
	if len(args) == 2 {
		archivePath = args[1]
	}

	_, err := sk.ArchiveSketch(context.Background(),
		&rpc.ArchiveSketchRequest{
			SketchPath:      sketchPath.String(),
			ArchivePath:     archivePath,
			IncludeBuildDir: includeBuildDir,
		})

	if err != nil {
		feedback.Errorf(tr("Error archiving: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
}
