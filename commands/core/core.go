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

package core

import (
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/common/formatter/pretty_print"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/bcmi-labs/arduino-cli/cores"
	"github.com/bcmi-labs/arduino-cli/cores/packageindex"
	"github.com/bcmi-labs/arduino-cli/pathutils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Init prepares the command.
func Init(rootCommand *cobra.Command) {
	rootCommand.AddCommand(command)
}

var command = &cobra.Command{
	Use:     "core",
	Short:   "Arduino Core operations.",
	Long:    "Arduino Core operations.",
	Example: "arduino core update-index # to update the package index file.",
}

// getInstalledCores gets the installed cores and puts them in the output struct.
func getInstalledCores(packageName string, cores *[]output.InstalledStuff) {
	getInstalledStuff(cores, configs.CoresFolder(packageName))
}

// getInstalledTools gets the installed tools and puts them in the output struct.
func getInstalledTools(packageName string, tools *[]output.InstalledStuff) {
	getInstalledStuff(tools, configs.ToolsFolder(packageName))
}

// getInstalledStuff is a generic procedure to get installed cores or tools and put them in an output struct.
func getInstalledStuff(stuff *[]output.InstalledStuff, folder pathutils.Path) {
	stuffHome, err := folder.Get()
	if err != nil {
		logrus.WithError(err).Warn("Cannot get default folder")
		return
	}
	stuffHomeFolder, err := os.Open(stuffHome)
	if err != nil {
		logrus.WithError(err).Warn("Cannot open default folder")
		return
	}
	defer stuffHomeFolder.Close()
	stuffFolders, err := stuffHomeFolder.Readdir(0)
	if err != nil {
		logrus.WithError(err).Warn("Cannot read into default folder")
		return
	}
	for _, stuffFolderInfo := range stuffFolders {
		if !stuffFolderInfo.IsDir() {
			continue
		}
		stuffName := stuffFolderInfo.Name()
		stuffFolder, err := os.Open(filepath.Join(stuffHome, stuffName))
		if err != nil {
			logrus.WithError(err).Warn("Cannot open inner directory")
			continue
		}
		defer stuffFolder.Close()
		versions, err := stuffFolder.Readdirnames(0)
		if err != nil {
			logrus.WithError(err).Warn("Cannot read into inner directory")
			continue
		}
		logrus.WithField("Name", stuffName).Info("Item added")
		*stuff = append(*stuff, output.InstalledStuff{
			Name:     stuffName,
			Versions: versions,
		})
	}
}

func getPackagesStatusContext() (*cores.StatusContext, error) {
	var index packageindex.Index
	err := packageindex.LoadIndex(&index)
	if err != nil {
		status, err := prettyPrints.CorruptedCoreIndexFix(index)
		if err != nil {
			return nil, err
		}
		return &status, nil
	}

	status := index.CreateStatusContext()
	return &status, nil
}
