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

package builder

import (
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/pkg/errors"
)

type HardwareLoader struct{}

func (s *HardwareLoader) Run(ctx *types.Context) error {
	if ctx.PackageManager == nil {
		pm := packagemanager.NewPackageManager(nil, nil, nil, nil)
		if errs := pm.LoadHardwareFromDirectories(ctx.HardwareDirs); len(errs) > 0 {
			// With the refactoring of the initialization step of the CLI we changed how
			// errors are returned when loading platforms and libraries, that meant returning a list of
			// errors instead of a single one to enhance the experience for the user.
			// I have no intention right now to start a refactoring of the legacy package too, so
			// here's this shitty solution for now.
			// When we're gonna refactor the legacy package this will be gone.
			return errors.WithStack(errs[0].Err())
		}
		ctx.PackageManager = pm
	}
	ctx.Hardware = ctx.PackageManager.Packages
	return nil
}
