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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/cmd/output"
	"github.com/bcmi-labs/arduino-cli/cmd/pretty_print"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/bcmi-labs/arduino-cli/cores"
	"github.com/spf13/cobra"
)

const (
	// CoreVersion represents the `arduino core` package version number.
	CoreVersion string = "0.1.0-alpha.preview"
)

var arduinoCoreCmd = &cobra.Command{
	Use:     "core",
	Short:   "Arduino Core operations",
	Long:    `Arduino Core operations`,
	Run:     executeCoreCommand,
	Example: `arduino core --update-index to update the package index file`,
}

// arduinoCoreVersionCmd represents the version command.
var arduinoCoreVersionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Shows version Number of arduino core package",
	Long:    `Shows version Number of arduino core package which is installed on your system.`,
	Run:     executeVersionCommand,
	Example: arduinoVersionCmd.Example,
}

var arduinoCoreListCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows the list of installed cores",
	Long: `Shows the list of installed cores. 
With -v tag (up to 2 times) can provide more verbose output.`,
	Run:     executeCoreListCommand,
	Example: `arduino core list -v for a medium verbosity level`,
}

var arduinoCoreDownloadCmd = &cobra.Command{
	Use:   "download [PACKAGER:ARCH[=VERSION]](S)",
	Short: "Downloads one or more cores and relative tool dependencies",
	Long:  `Downloads one or more cores and relative tool dependencies`,
	RunE:  executeCoreDownloadCommand,
	Example: `
arduino core download arduino:samd #to download latest version of arduino SAMD core.
arduino core download arduino:samd=1.6.9 #for the specific version (in this case 1.6.9)`,
}

var arduinoCoreInstallCmd = &cobra.Command{
	Use:   "install [PACKAGER:ARCH[=VERSION]](S)",
	Short: "Installs one or more cores and relative tool dependencies",
	Long:  `Installs one or more cores and relative tool dependencies`,
	RunE:  executeCoreInstallCommand,
	Example: `
arduino core install arduino:samd #to download latest version of arduino SAMD core.
arduino core installteele arduino:samd=1.6.9 #for the specific version (in this case 1.6.9)`,
}

func init() {
	versions[arduinoCoreCmd.Name()] = CoreVersion
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

	pkgs := output.InstalledPackageList{
		InstalledPackages: make([]output.InstalledPackage, 0, 10),
	}

	for _, file := range dirFiles {
		if !file.IsDir() {
			continue
		}
		packageName := file.Name()
		pkg := output.InstalledPackage{
			Name:           packageName,
			InstalledCores: make([]output.InstalledStuff, 0, 5),
			InstalledTools: make([]output.InstalledStuff, 0, 5),
		}
		getInstalledCores(packageName, &pkg.InstalledCores)
		getInstalledTools(packageName, &pkg.InstalledTools)
		pkgs.InstalledPackages = append(pkgs.InstalledPackages, pkg)
	}

	formatter.Print(pkgs)
}

func executeCoreDownloadCommand(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("No core specified for download command")
	}

	status, err := getPackagesStatusContext(GlobalFlags.Verbose)
	if err != nil {
		return nil
	}

	IDTuples := cores.ParseArgs(args)

	coresToDownload, toolsToDownload, failOutputs := status.Process(IDTuples)
	outputResults := output.CoreProcessResults{
		Cores: failOutputs,
		Tools: make([]output.ProcessResult, 0, 10),
	}
	downloads := make([]releases.DownloadItem, len(toolsToDownload))
	for i := range toolsToDownload {
		downloads[i] = toolsToDownload[i].DownloadItem
	}

	releases.ParallelDownload(downloads, true, "Downloaded", GlobalFlags.Verbose, &outputResults.Tools, "tool")
	downloads = make([]releases.DownloadItem, len(coresToDownload))
	for i := range coresToDownload {
		downloads[i] = coresToDownload[i].DownloadItem
	}
	releases.ParallelDownload(downloads, true, "Downloaded", GlobalFlags.Verbose, &outputResults.Cores, "core")

	formatter.Print(outputResults)
	return nil
}

func executeCoreInstallCommand(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("No core specified for download command")
	}

	status, err := getPackagesStatusContext(GlobalFlags.Verbose)
	if err != nil {
		return nil
	}

	IDTuples := cores.ParseArgs(args)
	coresToDownload, toolsToDownload, failOutputs := status.Process(IDTuples)
	failOutputsCount := len(failOutputs)
	outputResults := output.CoreProcessResults{
		Cores: failOutputs,
	}

	downloads := make([]releases.DownloadItem, len(toolsToDownload))
	for i := range toolsToDownload {
		downloads[i] = toolsToDownload[i].DownloadItem
	}
	releases.ParallelDownload(downloads, false, "Installed", GlobalFlags.Verbose, &outputResults.Tools, "tool")

	downloads = make([]releases.DownloadItem, len(coresToDownload))
	for i := range coresToDownload {
		downloads[i] = coresToDownload[i].DownloadItem
	}
	releases.ParallelDownload(downloads, false, "Installed", GlobalFlags.Verbose, &outputResults.Cores, "core")

	for i, item := range toolsToDownload {
		err = cores.InstallTool(item.Package, item.Name, item.Release)
		if err != nil {
			outputResults.Tools[i] = output.ProcessResult{
				ItemName: item.Name,
				Error:    err.Error(),
			}
		} else {
			toolRoot, err := common.GetDefaultToolsFolder(item.Package)
			if err != nil {
				formatter.PrintErrorMessage("Cannot get tool install path, try again.")
				return nil
			}
			outputResults.Tools[i].Path = filepath.Join(toolRoot, item.Name, item.Release.VersionName())
		}
	}

	for i, item := range coresToDownload {
		err = cores.Install(item.Package, item.Name, item.Release)
		if err != nil {
			outputResults.Cores[i+failOutputsCount] = output.ProcessResult{
				ItemName: item.Name,
				Status:   "",
				Error:    err.Error(),
			}
		} else {
			coreRoot, err := common.GetDefaultCoresFolder(item.Package)
			if err != nil {
				formatter.PrintErrorMessage("Cannot get core install path, try again.")
				return nil
			}
			outputResults.Cores[i+failOutputsCount].Path = filepath.Join(coreRoot, item.Name, item.Release.VersionName())
		}
	}

	formatter.Print(outputResults)
	return nil
}

// getInstalledCores gets the installed cores and puts them in the output struct.
func getInstalledCores(packageName string, cores *[]output.InstalledStuff) {
	getInstalledStuff(packageName, cores, common.GetDefaultCoresFolder)
}

// getInstalledTools gets the installed tools and puts them in the output struct.
func getInstalledTools(packageName string, tools *[]output.InstalledStuff) {
	getInstalledStuff(packageName, tools, common.GetDefaultToolsFolder)
}

// getInstalledStuff is a generic procedure to get installed cores or tools and put them in an output struct.
func getInstalledStuff(packageName string, stuff *[]output.InstalledStuff, startPathFunc func(string) (string, error)) {
	stuffHome, err := startPathFunc(packageName)
	if err != nil {
		return
	}
	stuffHomeFolder, err := os.Open(stuffHome)
	if err != nil {
		return
	}
	defer stuffHomeFolder.Close()
	stuffFolders, err := stuffHomeFolder.Readdir(0)
	if err != nil {
		return
	}
	for _, stuffFolderInfo := range stuffFolders {
		if !stuffFolderInfo.IsDir() {
			continue
		}
		stuffName := stuffFolderInfo.Name()
		stuffFolder, err := os.Open(filepath.Join(stuffHome, stuffName))
		if err != nil {
			continue
		}
		defer stuffFolder.Close()
		versions, err := stuffFolder.Readdirnames(0)
		if err != nil {
			continue
		}
		*stuff = append(*stuff, output.InstalledStuff{
			Name:     stuffName,
			Versions: versions,
		})
	}
}

func getPackagesStatusContext(verbosity int) (*cores.StatusContext, error) {
	var index cores.Index
	err := cores.LoadIndex(&index)
	if err != nil {
		status, err := prettyPrints.CorruptedCoreIndexFix(index, verbosity)
		if err != nil {
			return nil, err
		}
		return &status, nil
	}

	status := index.CreateStatusContext()
	return &status, nil
}
