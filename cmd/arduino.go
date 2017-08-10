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
	"fmt"
	"os"
	"os/user"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/cmd/output"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/spf13/cobra"
)

const (
	bashAutoCompletionFunction = `
    __arduino_autocomplete() 
    {
        case $(last_command) in
            arduino)
    		    opts="lib core help version"
    		    ;;
            arduino_lib)
    		    opts="install uninstall list search version --update-index"
    	        ;;			
			arduino_core)
			    opts="install uninstall list search version --update-index"
				;;
    		arduino_help)
    		    opts="lib core version"
    		    ;;
	    esac		  
    	if [[ ${cur} == " *" ]] ; then
            COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
            return 0
        fi
	    return 1
    }`

	// ArduinoVersion represents Arduino CLI version number.
	ArduinoVersion = "0.1.0-alpha.preview"
)

var versions = make(map[string]string)

// arduinoCmd represents the base command when called without any subcommands
var arduinoCmd = &cobra.Command{
	Use:   "arduino",
	Short: "Arduino CLI",
	Long:  "Arduino Create Command Line Interface (arduino-cli)",
	BashCompletionFunction: bashAutoCompletionFunction,
	PersistentPreRun:       arduinoPreRun,
	RunE:                   arduinoRun,
	Example: `arduino --config /path/to/my/config/file # sets the config file instead of passing parameters or using the default one.
arduino --generate-docs to generate the docs and autocompletion for the whole CLI.
arduino --home /new/arduino/home/folder`,
}

// arduinoVersionCmd represents the version command.
var arduinoVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version Number of arduino CLI components",
	Long:  `Shows version Number of arduino CLI components which are installed on your system.`,
	Run:   executeVersionCommand,
	Example: `arduino version     # for the versions of all components.
arduino lib version # for the version of the lib component.
arduino core version # for the version of the core component.`,
}

func init() {

	versions[arduinoCmd.Name()] = ArduinoVersion
	arduinoCmd.PersistentFlags().CountVarP(&GlobalFlags.Verbose, "verbose", "v", "enables verbose output (use more times for a higher level)")
	arduinoCmd.PersistentFlags().StringVar(&GlobalFlags.Format, "format", "invalid", "the output format, can be [text|json]")
	arduinoCmd.PersistentFlags().StringVar(&GlobalFlags.Home, "home", "", "the custom home (if not specified $HOME will be used)")

	arduinoCmd.Flags().StringVar(&rootCmdFlags.ConfigFile, "config", "", "config file (default is $HOME/.arduino.yaml)")
	arduinoCmd.Flags().BoolVar(&rootCmdFlags.GenerateDocs, "generate-docs", false, "generates the docs for the CLI and puts it in docs folder")
	arduinoCmd.AddCommand(arduinoVersionCmd)
}

func arduinoPreRun(cmd *cobra.Command, args []string) {
	if !formatter.IsSupported(GlobalFlags.Format) {
		GlobalFlags.Format = "text"
	}
	formatter.SetFormatter(GlobalFlags.Format)
	if !formatter.IsCurrentFormat("text") {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			formatter.PrintErrorMessage("Invalid Call : should show Help, but it is available only in TEXT mode")
		})
	}
	if GlobalFlags.Home == "" {
		usr, err := user.Current()
		if err != nil {
			common.RootDirPath = GlobalFlags.Home
		}
		common.RootDirPath = usr.HomeDir
	} else {
		common.RootDirPath = GlobalFlags.Home
	}
}

func arduinoRun(cmd *cobra.Command, args []string) error {
	if rootCmdFlags.GenerateDocs {
		errorText := ""
		err := cmd.GenBashCompletionFile("docs/bash_completions/arduino")
		if err != nil {
			errorText += fmt.Sprintln(err.Error())
		}
		err = generateManPages(cmd)
		if err != nil {
			errorText += fmt.Sprintln(err.Error())
		}
		if errorText != "" {
			formatter.PrintErrorMessage(errorText)
		}
		return nil
	}
	return cmd.Help()
}

// Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	err := arduinoCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func executeVersionCommand(cmd *cobra.Command, args []string) {
	versionPrint(versionsToPrint(cmd, true)...)
}

func versionsToPrint(cmd *cobra.Command, isRoot bool) []string {
	verToPrint := make([]string, 0, 10)
	if isRoot {
		verToPrint = append(verToPrint, cmd.Parent().Name())
	}

	if GlobalFlags.Verbose > 0 {
		siblings := findSiblings(cmd)
		//search version command in siblings children.
		for _, sibling := range siblings {
			for _, sibChild := range sibling.Commands() {
				//fmt.Println(sibling.Name(), " >", sibChild.Name())
				if sibChild.Name() == "version" {
					verToPrint = append(verToPrint, sibling.Name())
				} else {
					verToPrint = append(verToPrint, versionsToPrint(sibChild, false)...)
				}
			}
		}
	}

	return verToPrint
}

// versionPrint formats and prints the version of the specified command.
func versionPrint(commandNames ...string) {
	if len(commandNames) == 1 {
		verCommand := output.VersionResult{
			CommandName: commandNames[0],
			Version:     versions[commandNames[0]],
		}
		formatter.Print(verCommand)
	} else {
		verFullInfo := output.VersionFullInfo{
			Versions: make([]output.VersionResult, len(commandNames)),
		}

		for i, commandName := range commandNames {
			verFullInfo.Versions[i] = output.VersionResult{
				CommandName: commandName,
				Version:     versions[commandName],
			}
		}

		formatter.Print(verFullInfo)
	}
}

// findSiblings returns the array of the siblings of the specified command.
func findSiblings(cmd *cobra.Command) (siblings []*cobra.Command) {
	for _, childCmd := range cmd.Parent().Commands() {
		if childCmd.Name() != "version" {
			siblings = append(siblings, childCmd)
		}
	}
	return
}
