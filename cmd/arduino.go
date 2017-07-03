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

	"errors"

	"github.com/bcmi-labs/arduino-cli/common"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	bashAutoCompletionFunction = `
    __arduino_autocomplete() 
    {
        case $(last_command) in
            arduino_lib)
    		    opts="install uninstall list search version --update-index"
    	        ;;
			arduino_core)
			    opts="install uninstall list search version --update-index"
				;;
    		arduino_help)
    		    opts="lib core version"
    		    ;;
            arduino)
    		    opts="lib core help version"
    		    ;;
	    esac		  
    	if [[ ${cur} == " *" ]] ; then
            COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
            return 0
        fi
	    return 1
    }`

	// ArduinoVersion represents Arduino CLI version number.
	ArduinoVersion string = "0.0.1-pre-alpha"
)

// GlobalFlags represents flags available in all the program.
var GlobalFlags struct {
	Verbose int
	Format  string
}

// rootCmdFlags represent flags available to the root command.
var rootCmdFlags struct {
	ConfigFile string
}

// arduinoCmd represents the base command when called without any subcommands
var arduinoCmd = &cobra.Command{
	Use:   "arduino",
	Short: "Arduino CLI",
	Long:  "Arduino Create Command Line Interface (arduino-cli)",
	BashCompletionFunction: bashAutoCompletionFunction,
	PersistentPreRun:       arduinoPreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("")
	},
}

// arduinoVersionCmd represents the version command.
var arduinoVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version Number of arduino",
	Long:  `Shows version Number of arduino which is installed on your system.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Infof("arduino V. %s\n", ArduinoVersion)
		if GlobalFlags.Verbose > 0 {
			logrus.Infof("arduino V. %s\n", LibVersion)
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	arduinoCmd.PersistentFlags().CountVarP(&GlobalFlags.Verbose, "verbose", "v", "enables verbose output (use more times for a higher level)")
	arduinoCmd.PersistentFlags().StringVar(&GlobalFlags.Format, "format", "invalid", "the output format, can be [text|json]")
	arduinoCmd.Flags().StringVar(&rootCmdFlags.ConfigFile, "config", "", "config file (default is $HOME/.arduino.yaml)")

	arduinoCmd.AddCommand(arduinoVersionCmd)
}

func arduinoPreRun(cmd *cobra.Command, args []string) {
	_, formatterExists := common.Formatters[GlobalFlags.Format]
	if !formatterExists {
		GlobalFlags.Format = "text"
	}
	logrus.SetFormatter(common.Formatters[GlobalFlags.Format])
}

// Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	err := arduinoCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if rootCmdFlags.ConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(rootCmdFlags.ConfigFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			logrus.Warn("Error while searching for home directory for this user")
			if GlobalFlags.Verbose > 0 {
				logrus.WithField("error", err).Warnf(": %s\n", err.Error())
			}
			logrus.Warnln()
			os.Exit(1)
		}

		// Search config in home directory with name ".arduino-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".arduino")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Infoln("Using config file:", viper.ConfigFileUsed())
	}
}
