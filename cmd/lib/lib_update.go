// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package libCmd

import (
	"fmt"

	"github.com/bcmi-labs/arduino-cli/libraries"
	"github.com/spf13/cobra"
)

// updateCmd represents the lib list update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates the library index to latest version",
	Long:  `Updates the library index to latest version from downloads.arduino.cc repository.`,
	Run:   execUpdateListIndex,
}

func init() {
	LibListCmd.AddCommand(updateCmd)
}

func execUpdateListIndex(cmd *cobra.Command, args []string) {
	fmt.Print("Downloading libraries index file from download.arduino.cc... ")
	err := libraries.DownloadLibrariesFile()
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println("Cannot download index file.")
		return
	}
	fmt.Println("DONE")
}
