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

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const (
	cliVersion  string = "0.0.1-pre-alpha"
	libsVersion string = "0.0.1-pre-alpha"
)

// CliVersionCmd represents the version command.
var CliVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version Number of arduino",
	Long:  `Shows version Number of arduino which is installed on your system.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("arduino V. %s\n", cliVersion)
		if GlobalFlags.Verbose > 0 {
			fmt.Printf("arduino lib V. %s\n", libsVersion)
		}
	},
}

// LibVersionCmd represents the version command.
var LibVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version Number of arduino lib",
	Long:  `Shows version Number of arduino lib which is installed on your system.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("arduino lib V. %s\n", libsVersion)
	},
}

//TODO : maybe it is possible to autogenerate versions from ancestors, as I wrote in this function.
func ancestorsBreadcrumb(cmd *cobra.Command) string {
	ancestors := make([]string, 2, 2)
	cmd.VisitParents(func(ancestor *cobra.Command) {
		ancestors = append(ancestors, ancestor.Use)
	})
	//fmt.Println(ancestors)
	return strings.Trim(strings.Join(ancestors, " "), " ")
}

func init() {
	RootCmd.AddCommand(CliVersionCmd)
	LibRoot.AddCommand(LibVersionCmd)
}
