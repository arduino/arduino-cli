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
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-paths-helper"
)

type ToolsLoader struct{}

func (s *ToolsLoader) Run(ctx *types.Context) error {
	if ctx.CanUseCachedTools {
		return nil
	}

	builtinFolders := paths.NewPathList()
	if ctx.BuiltInToolsDirs != nil {
		builtinFolders = ctx.BuiltInToolsDirs
	}

	pm := ctx.PackageManager
	pm.LoadToolsFromBundleDirectories(builtinFolders)

	ctx.CanUseCachedTools = true
	ctx.AllTools = pm.GetAllInstalledToolsReleases()

	if ctx.TargetBoard != nil {
		requiredTools, err := pm.FindToolsRequiredForBoard(ctx.TargetBoard)
		if err != nil {
			return err
		}
		ctx.RequiredTools = requiredTools
	}

	return nil
}
