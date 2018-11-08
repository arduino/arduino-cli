/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package core

import (
	"io/ioutil"
	"net/url"
	"os"
	"path"

	"go.bug.st/downloader"

	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUpdateIndexCommand() *cobra.Command {
	updateIndexCommand := &cobra.Command{
		Use:     "update-index",
		Short:   "Updates the index of cores.",
		Long:    "Updates the index of cores to the latest version.",
		Example: "  " + commands.AppName + " core update-index",
		Args:    cobra.NoArgs,
		Run:     runUpdateIndexCommand,
	}
	return updateIndexCommand
}

func runUpdateIndexCommand(cmd *cobra.Command, args []string) {
	updateIndexes()
}

func updateIndexes() {
	for _, URL := range commands.Config.BoardManagerAdditionalUrls {
		updateIndex(URL)
	}
}

// TODO: This should be in packagemanager......
func updateIndex(URL *url.URL) {
	logrus.WithField("url", URL).Print("Updating index")

	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		formatter.PrintError(err, "Error creating temp file for download")
		os.Exit(commands.ErrGeneric)
	}
	if err := tmpFile.Close(); err != nil {
		formatter.PrintError(err, "Error creating temp file for download")
		os.Exit(commands.ErrGeneric)
	}
	tmp := paths.New(tmpFile.Name())
	defer tmp.Remove()

	d, err := downloader.Download(tmp.String(), URL.String())
	if err != nil {
		formatter.PrintError(err, "Error downloading index "+URL.String())
		os.Exit(commands.ErrNetwork)
	}
	indexDirPath := commands.Config.IndexesDir()
	coreIndexPath := indexDirPath.Join(path.Base(URL.Path))
	formatter.DownloadProgressBar(d, "Updating index: "+coreIndexPath.Base())
	if d.Error() != nil {
		formatter.PrintError(d.Error(), "Error downloading index "+URL.String())
		os.Exit(commands.ErrNetwork)
	}

	if _, err := packageindex.LoadIndex(tmp); err != nil {
		formatter.PrintError(err, "Invalid package index in "+URL.String())
		os.Exit(commands.ErrGeneric)
	}

	if err := indexDirPath.MkdirAll(); err != nil {
		formatter.PrintError(err, "Can't create data directory "+indexDirPath.String())
		os.Exit(commands.ErrGeneric)
	}

	if err := tmp.CopyTo(coreIndexPath); err != nil {
		formatter.PrintError(err, "Error saving downloaded index "+URL.String())
		os.Exit(commands.ErrGeneric)
	}
}
