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
	"errors"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
)

var tr = i18n.Tr

// The Arduino CLI run in command line mode needs only a single instance
// so we always use the same ID
const instanceID int32 = 1

// Create a new CoreInstance, returns an error if creation fails.
// It's meant to be be called only once so it panics if called multiple times.
// We prefer to panic instead of returning an error since we don't want
// to risk a developer calling it multiple times different config files.
func Create(configFile string) {
	inst := commands.GetInstance(instanceID)
	if inst != nil {
		panic("instance Create called multiple times")
	}

	_, err := commands.Create(&rpc.CreateRequest{
		ConfigFile: configFile,
	})
	if err != nil {
		feedback.Errorf(tr("Error creating instance: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
}

// Get returns the one and only CoreInstance used by the Arduino CLI
// when running in the command line mode.
func Get() *commands.CoreInstance {
	inst := commands.GetInstance(instanceID)
	if inst == nil {
		feedback.Errorf(tr("Arduino CLI instance doesn't exist"))
		os.Exit(errorcodes.ErrGeneric)
	}
	return inst
}

// Init initializes the command line instance by loading installed libraries and platforms.
// In case of loading failures prints a list of errors for each
// platform or library that we failed to load.
// Package and library indexes files are automatically updated if the
// CLI is run for the first time.
func Init() {
	errs := []error{}
	// In case the CLI is executed for the first time
	if err := FirstUpdate(); err != nil {
		feedback.Errorf(tr("Error initializing instance: %v"), err)
		return
	}

	downloadCallback := output.ProgressBar()
	taskCallback := output.TaskProgress()

	instance := Get()
	err := commands.Init(&rpc.InitRequest{
		Instance: instance.ToRPC(),
	}, func(res *rpc.InitResponse) {
		if st := res.GetError(); st != nil {
			errs = append(errs, errors.New(st.Message))
		}

		if progress := res.GetInitProgress(); progress != nil {
			if progress.DownloadProgress != nil {
				downloadCallback(progress.DownloadProgress)
			}
			if progress.TaskProgress != nil {
				taskCallback(progress.TaskProgress)
			}
		}
	})
	if err != nil {
		errs = append(errs, err)
	}
	for _, err := range errs {
		feedback.Errorf(tr("Error initializing instance: %v"), err)
	}
}

// FirstUpdate downloads libraries and packages indexes if they don't exist.
// This ideally is only executed the first time the CLI is run.
func FirstUpdate() error {
	instance := Get()
	// Gets the data directory to verify if library_index.json and package_index.json exist
	dataDir := paths.New(instance.Settings.GetString("directories.data"))

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
				Instance: instance.ToRPC(),
			},
			output.ProgressBar(),
		)
		if err != nil {
			return err
		}
	}

	// The package_index.json file doesn't exists, that means the CLI is run for the first time,
	// similarly to the library update we download that file and all the other package indexes
	// from additional_urls
	if packageIndex.NotExist() {
		_, err := commands.UpdateIndex(context.Background(),
			&rpc.UpdateIndexRequest{
				Instance: instance.ToRPC(),
			},
			output.ProgressBar(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
