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

package upload

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/upload"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbn           string
	port           string
	verbose        bool
	verify         bool
	importDir      string
	programmer     string
	burnBootloader bool
)

// NewCommand created a new `upload` command
func NewCommand() *cobra.Command {
	uploadCommand := &cobra.Command{
		Use:     "upload",
		Short:   "Upload Arduino sketches.",
		Long:    "Upload Arduino sketches.",
		Example: "  " + os.Args[0] + " upload /home/user/Arduino/MySketch",
		Args:    cobra.MaximumNArgs(1),
		Run:     run,
	}

	uploadCommand.Flags().StringVarP(&fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:avr:uno")
	uploadCommand.Flags().StringVarP(&port, "port", "p", "", "Upload port, e.g.: COM10 or /dev/ttyACM0")
	uploadCommand.Flags().StringVarP(&importDir, "input-dir", "", "", "Directory containing binaries to upload.")
	uploadCommand.Flags().BoolVarP(&verify, "verify", "t", false, "Verify uploaded binary after the upload.")
	uploadCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, "Optional, turns on verbose mode.")
	uploadCommand.Flags().StringVarP(&programmer, "programmer", "P", "", "Optional, use the specified programmer to upload or 'list' to list supported programmers.")

	return uploadCommand
}

func run(command *cobra.Command, args []string) {
	instance, err := instance.CreateInstance()
	if err != nil {
		feedback.Errorf("Error during Upload: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	if programmer == "list" {
		resp, err := upload.ListProgrammersAvailableForUpload(context.Background(), &rpc.ListProgrammersAvailableForUploadReq{
			Instance: instance,
			Fqbn:     fqbn,
		})
		if err != nil {
			feedback.Errorf("Error listing programmers: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
		feedback.PrintResult(&programmersList{
			Programmers: resp.GetProgrammers(),
		})
		os.Exit(0)
	}

	var path *paths.Path
	if len(args) > 0 {
		path = paths.New(args[0])
	}
	sketchPath := initSketchPath(path)

	if burnBootloader {
		if _, err := upload.Upload(context.Background(), &rpc.UploadReq{
			Instance:   instance,
			Fqbn:       fqbn,
			SketchPath: sketchPath.String(),
			Port:       port,
			Verbose:    verbose,
			Verify:     verify,
			ImportDir:  importDir,
			Programmer: programmer,
		}, os.Stdout, os.Stderr); err != nil {
			feedback.Errorf("Error during Upload: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
		os.Exit(0)
	}

	if _, err := upload.Upload(context.Background(), &rpc.UploadReq{
		Instance:   instance,
		Fqbn:       fqbn,
		SketchPath: sketchPath.String(),
		Port:       port,
		Verbose:    verbose,
		Verify:     verify,
		ImportDir:  importDir,
		Programmer: programmer,
	}, os.Stdout, os.Stderr); err != nil {
		feedback.Errorf("Error during Upload: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
}

// initSketchPath returns the current working directory
func initSketchPath(sketchPath *paths.Path) *paths.Path {
	if sketchPath != nil {
		return sketchPath
	}

	wd, err := paths.Getwd()
	if err != nil {
		feedback.Errorf("Couldn't get current working directory: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	logrus.Infof("Reading sketch from dir: %s", wd)
	return wd
}

type programmersList struct {
	Programmers []*rpc.Programmer
}

func (p *programmersList) Data() interface{} {
	return p.Programmers
}

func (p *programmersList) String() string {
	t := table.New()
	t.SetHeader("ID", "Programmer Name", "Platform")
	for _, prog := range p.Programmers {
		t.AddRow(prog.GetId(), prog.GetName(), prog.GetPlatform())
	}
	return t.Render()
}
