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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package cmd

import (
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/cmd/pretty_print"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/spf13/cobra"
)

var arduinoCoreCmd = &cobra.Command{
	Use:   "core",
	Short: "Arduino Core operations",
	Long:  `Arduino Core operations`,
	Run:   executeCoreCommand,
}

var arduinoCoreListCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows the list of installed cores",
	Long: `Shows the list of installed cores. 
With -v tag (up to 2 times) can provide more verbose output.`,
	Run: executeCoreListCommand,
}

func init() {
	arduinoCmd.AddCommand(arduinoCoreCmd)
	arduinoCoreCmd.AddCommand(arduinoCoreListCmd)

	arduinoCoreCmd.Flags().BoolVar(&arduinoCoreFlags.updateIndex, "update-index", false, "Updates the index of cores to the latest version")
}

func executeCoreCommand(cmd *cobra.Command, args []string) {
	if arduinoCoreFlags.updateIndex {
		common.ExecUpdateIndex(prettyPrints.DownloadCoreFileIndex(), GlobalFlags.Verbose)
	} else {
		cmd.Help()
	}
}

func executeCoreListCommand(cmd *cobra.Command, args []string) {
	pkgHome, err := common.GetDefaultPkgFolder()
	if err != nil {
		formatter.PrintError(err)
		return
	}

	dir, err := os.Open(pkgHome)
	if err != nil {
		formatter.PrintErrorMessage("Cannot open packages folder")
		return
	}
	defer dir.Close()

	dirFiles, err := dir.Readdir(0)
	if err != nil {
		formatter.PrintErrorMessage("Cannot read into packages folder")
		return
	}

	coreMap := make(map[string]map[string][]string, 10)
	toolMap := make(map[string]map[string][]string, 10)

	for _, file := range dirFiles {
		if file.IsDir() {
			packageName := file.Name()
			coreHome, err := common.GetDefaultCoresFolder(packageName)
			if err == nil {
				coreHomeFolder, err := os.Open(coreHome)
				if err == nil {
					defer coreHomeFolder.Close()
					cores, err := coreHomeFolder.Readdir(0)
					if err == nil {
						for _, core := range cores {
							if core.IsDir() {
								coreName := core.Name()
								coreDir, err := os.Open(filepath.Join(coreHome, coreName))
								if err == nil {
									defer coreDir.Close()
									versions, err := coreDir.Readdirnames(0)
									if err == nil {
										coreMap[packageName] = make(map[string][]string, len(versions))
										coreMap[packageName][coreName] = versions
									} //else skip
								} //else skip
							} //else skip
						}
					} //else skip
				} //else skip
			} //else skip
			toolsHome, err := common.GetDefaultToolsFolder(packageName)
			if err == nil {
				toolsHomeFolder, err := os.Open(toolsHome)
				if err == nil {
					defer toolsHomeFolder.Close()
					toolFolders, err := toolsHomeFolder.Readdir(0)
					if err == nil {
						for _, toolFolderInfo := range toolFolders {
							if toolFolderInfo.IsDir() {
								toolName := toolFolderInfo.Name()
								toolFolder, err := os.Open(filepath.Join(toolsHome, toolName))
								if err == nil {
									defer toolFolder.Close()
									versions, err := toolFolder.Readdirnames(0)
									if err == nil {
										toolMap[packageName] = make(map[string][]string)
										toolMap[packageName][toolName] = versions
									}
								} //else skip
							} //else skip
						}
					} //else skip
				} //else skip
			} //else skip
		} //else skip
	}
}
