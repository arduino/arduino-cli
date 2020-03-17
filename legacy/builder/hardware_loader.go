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
		if err := pm.LoadHardwareFromDirectories(ctx.HardwareDirs); err != nil {
			return errors.WithStack(err)
		}
		ctx.PackageManager = pm
	}
	ctx.Hardware = ctx.PackageManager.Packages
	return nil
}
