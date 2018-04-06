/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package builder

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-builder/types"
)

type ToolsLoader struct{}

func (s *ToolsLoader) Run(ctx *types.Context) error {
	folders := []string{}
	builtinFolders := []string{}

	if ctx.BuiltInToolsFolders != nil || len(ctx.BuiltInLibrariesFolders) == 0 {
		folders = ctx.ToolsFolders
		builtinFolders = ctx.BuiltInToolsFolders
	} else {
		// Auto-detect built-in tools folders (for arduino-builder backward compatibility)
		// this is a deprecated feature and will be removed in the future
		builtinHardwareFolder, err := filepath.Abs(filepath.Join(ctx.BuiltInLibrariesFolders[0], ".."))
		if err != nil {
			fmt.Println("Error detecting ")
		}

		builtinFolders = []string{}
		for _, folder := range ctx.ToolsFolders {
			if !strings.Contains(folder, builtinHardwareFolder) {
				folders = append(folders, folder)
			} else {
				builtinFolders = append(builtinFolders, folder)
			}
		}
	}

	pm := ctx.PackageManager
	pm.LoadToolsFromBundleDirectories(builtinFolders)

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
