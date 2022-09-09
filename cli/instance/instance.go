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
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
)

var tr = i18n.Tr

// CreateAndInit return a new initialized instance.
// If Create fails the CLI prints an error and exits since
// to execute further operations a valid Instance is mandatory.
// If Init returns errors they're printed only.
func CreateAndInit() *rpc.Instance {
	inst, _ := CreateAndInitWithProfile("", nil)
	return inst
}

// CreateAndInitWithProfile returns a new initialized instance using the given profile of the given sketch.
// If Create fails the CLI prints an error and exits since to execute further operations a valid Instance is mandatory.
// If Init returns errors they're printed only.
func CreateAndInitWithProfile(profileName string, sketchPath *paths.Path) (*rpc.Instance, *rpc.Profile) {
	instance, err := Create()
	if err != nil {
		feedback.Errorf(tr("Error creating instance: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
	profile, errs := InitWithProfile(instance, profileName, sketchPath)
	for _, err := range errs {
		feedback.Errorf(tr("Error initializing instance: %v"), err)
	}
	return instance, profile
}

// Create and return a new Instance.
func Create() (*rpc.Instance, error) {
	res, err := commands.Create(&rpc.CreateRequest{})
	if err != nil {
		return nil, err
	}
	return res.Instance, nil
}

// Init initializes instance by loading installed libraries and platforms.
// In case of loading failures return a list of errors for each
// platform or library that we failed to load.
// Package and library indexes files are automatically updated if the
// CLI is run for the first time.
func Init(instance *rpc.Instance) []error {
	_, errs := InitWithProfile(instance, "", nil)
	return errs
}

// InitWithProfile initializes instance by loading libraries and platforms specified in the given profile of the given sketch.
// In case of loading failures return a list of errors for each platform or library that we failed to load.
// Required Package and library indexes files are automatically downloaded.
func InitWithProfile(instance *rpc.Instance, profileName string, sketchPath *paths.Path) (*rpc.Profile, []error) {
	errs := []error{}

	// In case the CLI is executed for the first time
	if err := FirstUpdate(instance); err != nil {
		return nil, append(errs, err)
	}

	downloadCallback := output.ProgressBar()
	taskCallback := output.TaskProgress()

	initReq := &rpc.InitRequest{Instance: instance}
	if sketchPath != nil {
		initReq.SketchPath = sketchPath.String()
		initReq.Profile = profileName
	}
	var profile *rpc.Profile
	err := commands.Init(initReq, func(res *rpc.InitResponse) {
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

		if p := res.GetProfile(); p != nil {
			profile = p
		}
	})
	if err != nil {
		errs = append(errs, err)
	}

	return profile, errs
}

// FirstUpdate downloads libraries and packages indexes if they don't exist.
// This ideally is only executed the first time the CLI is run.
func FirstUpdate(instance *rpc.Instance) error {
	// Gets the data directory to verify if library_index.json and package_index.json exist
	dataDir := configuration.DataDir(configuration.Settings)

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
			return err
		}
	}

	// The package_index.json file doesn't exists, that means the CLI is run for the first time,
	// similarly to the library update we download that file and all the other package indexes
	// from additional_urls
	if packageIndex.NotExist() {
		err := commands.UpdateIndex(context.Background(),
			&rpc.UpdateIndexRequest{
				Instance:                   instance,
				IgnoreCustomPackageIndexes: true,
			},
			output.ProgressBar(),
			output.PrintErrorFromDownloadResult(tr("Error updating index")))
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateInstanceAndRunFirstUpdate creates an instance and runs `FirstUpdate`.
// It is mandatory for all `update-index` commands to call this
func CreateInstanceAndRunFirstUpdate() *rpc.Instance {
	// We don't initialize any CoreInstance when updating indexes since we don't need to.
	// Also meaningless errors might be returned when calling this command with --additional-urls
	// since the CLI would be searching for a corresponding file for the additional urls set
	// as argument but none would be obviously found.
	inst, status := Create()
	if status != nil {
		feedback.Errorf(tr("Error creating instance: %v"), status)
		os.Exit(errorcodes.ErrGeneric)
	}

	// In case this is the first time the CLI is run we need to update indexes
	// to make it work correctly, we must do this explicitly in this command since
	// we must use instance.Create instead of instance.CreateAndInit for the
	// reason stated above.
	if err := FirstUpdate(inst); err != nil {
		feedback.Errorf(tr("Error updating indexes: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
	return inst
}
