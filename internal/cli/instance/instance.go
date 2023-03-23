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
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
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
	instance, err := create()
	if err != nil {
		feedback.Fatal(tr("Error creating instance: %v", err), feedback.ErrGeneric)
	}
	profile := InitWithProfile(instance, profileName, sketchPath)
	return instance, profile
}

// create and return a new Instance.
func create() (*rpc.Instance, error) {
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
func Init(instance *rpc.Instance) {
	InitWithProfile(instance, "", nil)
}

// InitWithProfile initializes instance by loading libraries and platforms specified in the given profile of the given sketch.
// In case of loading failures return a list of errors for each platform or library that we failed to load.
// Required Package and library indexes files are automatically downloaded.
func InitWithProfile(instance *rpc.Instance, profileName string, sketchPath *paths.Path) *rpc.Profile {
	downloadCallback := feedback.ProgressBar()
	taskCallback := feedback.TaskProgress()

	initReq := &rpc.InitRequest{Instance: instance}
	if sketchPath != nil {
		initReq.SketchPath = sketchPath.String()
		initReq.Profile = profileName
	}
	var profile *rpc.Profile
	err := commands.Init(initReq, func(res *rpc.InitResponse) {
		if st := res.GetError(); st != nil {
			feedback.Warning(tr("Error initializing instance: %v", st.Message))
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
		feedback.Warning(tr("Error initializing instance: %v", err))
	}

	return profile
}
