/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
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
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package lib

import (
	"os"

	paths "github.com/arduino/go-paths-helper"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/configs"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// InitCommand prepares the command.
func InitCommand() *cobra.Command {
	libCommand := &cobra.Command{
		Use:   "lib",
		Short: "Arduino commands about libraries.",
		Long:  "Arduino commands about libraries.",
		Example: "" +
			"arduino lib install AudioZero\n" +
			"arduino lib update-index",
	}
	libCommand.AddCommand(initDownloadCommand())
	libCommand.AddCommand(initInstallCommand())
	libCommand.AddCommand(initListCommand())
	libCommand.AddCommand(initSearchCommand())
	libCommand.AddCommand(initUninstallCommand())
	libCommand.AddCommand(initUpdateIndexCommand())
	return libCommand
}

func getLibraryManager() *librariesmanager.LibrariesManager {
	logrus.Info("Starting libraries manager")
	pm := commands.InitPackageManager()
	lm := librariesmanager.NewLibraryManager()

	// Add IDE builtin libraries dir
	if bundledLibsDir := configs.IDEBundledLibrariesDir(); bundledLibsDir != nil {
		lm.AddLibrariesDir(bundledLibsDir, libraries.IDEBuiltIn)
	}

	// Add sketchbook libraries dir
	if libHome, err := configs.LibrariesFolder.Get(); err != nil {
		formatter.PrintError(err, "Cannot get libraries folder.")
		os.Exit(commands.ErrCoreConfig)
	} else {
		lm.AddLibrariesDir(paths.New(libHome), libraries.Sketchbook)
	}

	// Add libraries dirs from installed platforms
	for _, targetPackage := range pm.GetPackages().Packages {
		for _, platform := range targetPackage.Platforms {
			if platformRelease := platform.GetInstalled(); platformRelease != nil {
				lm.AddPlatformReleaseLibrariesDir(platformRelease, libraries.PlatformBuiltIn)
			}
		}
	}

	if err := lm.LoadIndex(); err != nil {
		logrus.WithError(err).Warn("Error during libraries index loading, try to download it again")
		updateIndex()
	}
	if err := lm.LoadIndex(); err != nil {
		logrus.WithError(err).Error("Error during libraries index loading")
		formatter.PrintError(err, "Error loading libraries index")
		os.Exit(commands.ErrGeneric)
	}
	return lm
}
