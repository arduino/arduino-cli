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

package instance

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateAndInit return a new initialized instance.
// If Create fails the CLI prints an error and exits since
// to execute further operations a valid Instance is mandatory.
// If Init returns errors they're printed only.
func CreateAndInit() *rpc.Instance {
	instance, err := Create()
	if err != nil {
		feedback.Errorf("Error creating instance: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	for _, err := range Init(instance) {
		feedback.Errorf("Error initializing instance: %v", err)
	}
	return instance
}

// Create and return a new Instance.
func Create() (*rpc.Instance, *status.Status) {
	res, err := commands.Create(&rpc.CreateRequest{})
	if err != nil {
		return nil, err
	}
	return res.Instance, nil
}

// Init initializes instance by loading installed libraries and platforms.
// In case of loading failures return a list gRPC Status errors for each
// platform or library that we failed to load.
// Package and library indexes files are automatically updated if the
// CLI is run for the first time.
func Init(instance *rpc.Instance) []*status.Status {
	errs := []*status.Status{}

	// In case the CLI is executed for the first time
	if err := FirstUpdate(instance); err != nil {
		return append(errs, err)
	}

	initChan, err := commands.Init(&rpc.InitRequest{
		Instance: instance,
	})
	if err != nil {
		return append(errs, err)
	}

	downloadCallback := output.ProgressBar()
	taskCallback := output.TaskProgress()

	for res := range initChan {
		if err := res.GetError(); err != nil {
			errs = append(errs, status.FromProto(err))
		}

		if progress := res.GetInitProgress(); progress != nil {
			if progress.DownloadProgress != nil {
				downloadCallback(progress.DownloadProgress)
			}
			if progress.TaskProgress != nil {
				taskCallback(progress.TaskProgress)
			}
		}
	}

	return errs
}

// FirstUpdate downloads libraries and packages indexes if they don't exist.
// This ideally is only executed the first time the CLI is run.
func FirstUpdate(instance *rpc.Instance) *status.Status {
	// Gets the data directory to verify if library_index.json and package_index.json exist
	dataDir := paths.New(configuration.Settings.GetString("directories.data"))

	libraryIndex := dataDir.Join("library_index.json")
	packageIndex := dataDir.Join("package_index.json")

	if libraryIndex.Exist() && packageIndex.Exist() {
		return nil
	}

	// The library_index.json file doesn't exists, that means the CLI is run for the first time
	// so we proceed with the first update that downloads the file
	if libraryIndex.NotExist() {
		err := commands.UpdateLibrariesIndex(context.Background(),
			&rpc.UpdateLibrariesIndexRequest{
				Instance: instance,
			},
			output.ProgressBar(),
		)
		if err != nil {
			return status.Newf(codes.FailedPrecondition, err.Error())
		}
	}

	// The package_index.json file doesn't exists, that means the CLI is run for the first time,
	// similarly to the library update we download that file and all the other package indexes
	// from additional_urls
	if packageIndex.NotExist() {
		_, err := commands.UpdateIndex(context.Background(),
			&rpc.UpdateIndexRequest{
				Instance: instance,
			},
			output.ProgressBar(),
		)
		if err != nil {
			return status.Newf(codes.FailedPrecondition, err.Error())
		}
	}

	return nil
}
