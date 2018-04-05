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

	"github.com/bcmi-labs/arduino-cli/commands"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	command.AddCommand(updateIndexCommand)
}

var updateIndexCommand = &cobra.Command{
	Use:     "update-index",
	Short:   "Updates the index of cores.",
	Long:    "Updates the index of cores to the latest version.",
	Example: "arduino core update-index",
	Args:    cobra.NoArgs,
	Run:     runUpdateIndexCommand,
}

func runUpdateIndexCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Updating package index")

	failed := false
	for _, URL := range configs.BoardManagerAdditionalUrls {
		formatter.Print("Downloading package index from " + URL.String())
		coreIndexPath := configs.IndexPathFromURL(URL)
		if err := common.DownloadIndex(coreIndexPath, URL); err != nil {
			formatter.PrintError(err, "Can't download index file.")
			failed = true
		}
	}

	if failed {
		formatter.PrintErrorMessage("Error downloading package index, check your network connection")
		os.Exit(commands.ErrNetwork)
	}
	formatter.Print("Download completed.")
}
