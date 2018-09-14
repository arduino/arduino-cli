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

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/go-paths-helper"
	"github.com/cavaliercoder/grab"
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
	for _, URL := range configs.BoardManagerAdditionalUrls {
		updateIndex(URL)
	}
}

// TODO: This should be in packagemanager......
func updateIndex(URL *url.URL) {
	logrus.WithField("url", URL).Print("Updating index")
	indexDirPath := commands.Config.IndexesDir()
	coreIndexPath := indexDirPath.Join(path.Base(URL.Path))

	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		formatter.PrintError(err, "Error creating temp file for download")
		os.Exit(commands.ErrGeneric)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	req, err := grab.NewRequest(tmpFile.Name(), URL.String())
	if err != nil {
		formatter.PrintError(err, "Error downloading index "+URL.String())
		os.Exit(commands.ErrNetwork)
	}
	client := grab.NewClient()
	resp := client.Do(req)
	formatter.DownloadProgressBar(resp, "Updating index: "+coreIndexPath.Base())
	if resp.Err() != nil {
		formatter.PrintError(resp.Err(), "Error downloading index "+URL.String())
		os.Exit(commands.ErrNetwork)
	}

	if err := indexDirPath.MkdirAll(); err != nil {
		formatter.PrintError(err, "Can't create data directory "+indexDirPath.String())
		os.Exit(commands.ErrGeneric)
	}

	if err := paths.New(tmpFile.Name()).CopyTo(coreIndexPath); err != nil {
		formatter.PrintError(err, "Error saving downloaded index "+URL.String())
		os.Exit(commands.ErrGeneric)
	}
}
