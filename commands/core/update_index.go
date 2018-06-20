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
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/arduino/go-paths-helper"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/cavaliercoder/grab"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUpdateIndexCommand() *cobra.Command {
	updateIndexCommand := &cobra.Command{
		Use:     "update-index",
		Short:   "Updates the index of cores.",
		Long:    "Updates the index of cores to the latest version.",
		Example: "arduino core update-index",
		Args:    cobra.NoArgs,
		Run:     runUpdateIndexCommand,
	}
	return updateIndexCommand
}

func runUpdateIndexCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Updating package index")
	updateIndexes()
}

func updateIndexes() {
	for _, URL := range configs.BoardManagerAdditionalUrls {
		updateIndex(URL)
	}
}

func updateIndex(URL *url.URL) {
	coreIndexPath, err := configs.IndexPathFromURL(URL).Get()
	if err != nil {
		formatter.PrintError(err, "Error getting index path for "+URL.String())
		os.Exit(commands.ErrGeneric)
	}

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
	formatter.DownloadProgressBar(resp, "Updating index: "+filepath.Base(coreIndexPath))
	if resp.Err() != nil {
		formatter.PrintError(resp.Err(), "Error downloading index "+URL.String())
		os.Exit(commands.ErrNetwork)
	}

	if err := paths.New(tmpFile.Name()).CopyTo(paths.New(coreIndexPath)); err != nil {
		formatter.PrintError(err, "Error saving downloaded index "+URL.String())
		os.Exit(commands.ErrGeneric)
	}
}
