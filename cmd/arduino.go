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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/mitchellh/go-homedir"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/spf13/cobra"
	"github.com/theherk/viper"

	"github.com/bcmi-labs/arduino-cli/auth"
	"github.com/bcmi-labs/arduino-cli/common"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/cmd/output"
	"github.com/bcmi-labs/arduino-cli/configs"
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

// ArduinoCmd represents the base command when called without any subcommands
var ArduinoCmd = &cobra.Command{
	Use:   "arduino",
	Short: "Arduino CLI",
	Long:  "Arduino Create Command Line Interface (arduino-cli)",
	BashCompletionFunction: bashAutoCompletionFunction,
	PersistentPreRun:       arduinoPreRun,
	Run:                    arduinoRun,
	Example:                `arduino --generate-docs to generate the docs and autocompletion for the whole CLI.`,
}

// arduinoVersionCmd represents the version command.
var arduinoVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version Number of arduino CLI components",
	Long:  `Shows version Number of arduino CLI components which are installed on your system.`,
	Run:   executeVersionCommand,
	Example: `arduino version     # for the main component version.
arduino lib version # for the version of the lib component.
arduino core version # for the version of the core component.`,
}

var arduinoLoginCmd = &cobra.Command{
	Use:   `login [--user USER --password PASSWORD | --user USER`,
	Short: `create default credentials for an Arduino Create Session`,
	Long:  `create default credentials for an Arduino Create Session`,
	Example: `arduino login                          # Asks all credentials.
arduino login --user myUser --password MySecretPassword
arduino login --user myUser --password # Asks to write the password inside the command instead of having it in clear.`,
	Run: executeLoginCommand,
}

var testing = false

// ErrLogrus represents the logrus instance, which has the role to
// log all non info messages.
var ErrLogrus = logrus.New()

func init() {
	versions[ArduinoCmd.Name()] = ArduinoVersion
	InitFlags()
	InitCommands()
}

// TestInit creates an initialization for tests
func TestInit() {
	InitFlags()
	InitCommands()
	InitConfigs()

	cobra.OnInitialize(func() {
		viper.SetConfigFile("./test-config.yml")
	})

	testing = true
}

// InitFlags reinitialize flags (useful for testing too)
func InitFlags() {
	ArduinoCmd.ResetFlags()
	arduinoVersionCmd.ResetFlags()

	arduinoLibCmd.ResetFlags()
	arduinoLibInstallCmd.ResetFlags()
	arduinoLibDownloadCmd.ResetFlags()
	arduinoLibListCmd.ResetFlags()
	arduinoLibSearchCmd.ResetFlags()
	arduinoLibUninstallCmd.ResetFlags()
	arduinoLibVersionCmd.ResetFlags()

	arduinoCoreCmd.ResetFlags()
	arduinoCoreDownloadCmd.ResetFlags()
	arduinoCoreInstallCmd.ResetFlags()
	arduinoCoreListCmd.ResetFlags()
	arduinoCoreVersionCmd.ResetFlags()

	arduinoConfigCmd.ResetFlags()
	arduinoConfigInitCmd.ResetFlags()

	arduinoBoardAttachCmd.ResetFlags()
	arduinoBoardListCmd.ResetFlags()

	arduinoSketchSyncCmd.ResetFlags()
	arduinoLoginCmd.ResetFlags()

	ArduinoCmd.PersistentFlags().BoolVar(&GlobalFlags.Debug, "debug", false, "enables debug output (super verbose, used to debug the CLI)")

	ArduinoCmd.PersistentFlags().StringVar(&GlobalFlags.Format, "format", "text", "the output format, can be [text|json]")

	ArduinoCmd.PersistentFlags().StringVar(&configs.FileLocation, "config-file", configs.FileLocation, "the custom config file (if not specified ./.cli-config.yml will be used)")

	ArduinoCmd.Flags().BoolVar(&rootCmdFlags.GenerateDocs, "generate-docs", false, "generates the docs for the CLI and puts it in docs folder")

	arduinoLibCmd.Flags().BoolVar(&arduinoLibFlags.updateIndex, "update-index", false, "Updates the libraries index")

	arduinoLibSearchCmd.Flags().BoolVar(&arduinoLibSearchFlags.Names, "names", false, "Show library names only")

	arduinoCoreCmd.Flags().BoolVar(&arduinoCoreFlags.updateIndex, "update-index", false, "Updates the index of cores to the latest version")

	arduinoConfigInitCmd.Flags().BoolVar(&arduinoConfigInitFlags.Default, "default", false, "If omitted, ask questions to the user about setting configuration properties, otherwise use default configuration")
	arduinoConfigInitCmd.Flags().StringVar(&arduinoConfigInitFlags.Location, "save-as", configs.FileLocation, "Sets where to save the configuration file [default is ./.cli-config.yml]")

	arduinoBoardListCmd.Flags().StringVar(&arduinoBoardListFlags.SearchTimeout, "timeout", "5s", "The timeout of the search of connected devices, try to high it if your board is not found (e.g. to 10s)")

	arduinoBoardAttachCmd.Flags().StringVar(&arduinoBoardAttachFlags.BoardURI, "board", "", "The URI of the board to connect")
	arduinoBoardAttachCmd.Flags().StringVar(&arduinoBoardAttachFlags.BoardFlavour, "flavour", "default", "The Name of the CPU flavour, it is required for some boards (e.g. Arduino Nano)")
	arduinoBoardAttachCmd.Flags().StringVar(&arduinoBoardAttachFlags.SketchName, "sketch", "", "The Name of the sketch to attach the board to")
	arduinoBoardAttachCmd.Flags().StringVar(&arduinoBoardAttachFlags.SearchTimeout, "timeout", "5s", "The timeout of the search of connected devices, try to high it if your board is not found (e.g. to 10s)")

	arduinoSketchSyncCmd.Flags().StringVar(&arduinoSketchSyncFlags.Priority, "conflict-policy", "skip", "The decision made by default on conflicting sketches. Can be push-local, pull-remote, skip, ask-once, ask-always.")
	arduinoLoginCmd.Flags().StringVarP(&arduinoLoginFlags.User, "user", "u", "", "The username used to log in")
	arduinoLoginCmd.Flags().StringVarP(&arduinoLoginFlags.Password, "password", "p", "", "The username used to log in")
}

// InitCommands reinitialize commands (useful for testing too)
func InitCommands() {
	ArduinoCmd.ResetCommands()
	arduinoLibCmd.ResetCommands()
	arduinoCoreCmd.ResetCommands()
	arduinoConfigCmd.ResetCommands()
	arduinoBoardCmd.ResetCommands()
	arduinoSketchCmd.ResetCommands()

	ArduinoCmd.AddCommand(arduinoVersionCmd, arduinoLibCmd, arduinoCoreCmd, arduinoConfigCmd,
		arduinoBoardCmd, arduinoSketchCmd, arduinoLoginCmd)

	arduinoLibCmd.AddCommand(arduinoLibInstallCmd, arduinoLibUninstallCmd, arduinoLibSearchCmd,
		arduinoLibVersionCmd, arduinoLibListCmd, arduinoLibDownloadCmd)

	arduinoCoreCmd.AddCommand(arduinoCoreListCmd, arduinoCoreDownloadCmd, arduinoCoreVersionCmd,
		arduinoCoreInstallCmd)

	arduinoConfigCmd.AddCommand(arduinoConfigInitCmd)

	arduinoBoardCmd.AddCommand(arduinoBoardListCmd, arduinoBoardAttachCmd)

	arduinoSketchCmd.AddCommand(arduinoSketchSyncCmd)
}

// InitConfigs initializes the configuration from the specified file.
func InitConfigs() {
	logrus.Info("Initiating configuration")
	c, err := configs.Unserialize(configs.FileLocation)
	if err != nil {
		logrus.WithError(err).Warn("Did not manage to get config file, using default configuration")
		GlobalFlags.Configs = configs.Default()
	}
	if configs.Bundled() {
		logrus.Info("CLI is bundled into the IDE")
		err := configs.UnserializeFromIDEPreferences(&c)
		if err != nil {
			logrus.WithError(err).Warn("Did not manage to get config file of IDE, using default configuration")
			GlobalFlags.Configs = configs.Default()
		}
	} else {
		logrus.Info("CLI is not bundled into the IDE")
	}
	logrus.Info("Configuration set")
	GlobalFlags.Configs = c
	common.ArduinoDataFolder = GlobalFlags.Configs.ArduinoDataFolder
	common.ArduinoIDEFolder = configs.ArduinoIDEFolder
	common.SketchbookFolder = GlobalFlags.Configs.SketchbookPath
}

// IgnoreConfigs is used in tests to ignore the config file.
func IgnoreConfigs() {
	logrus.Info("Ignoring configurations and using always default ones")
	GlobalFlags.Configs = configs.Default()
}

func arduinoPreRun(cmd *cobra.Command, args []string) {
	// Reset logrus if debug flag changed
	if !GlobalFlags.Debug { // discard logrus output if no debug
		logrus.SetOutput(ioutil.Discard) // for standard logger
	} else { // else print on stderr
		ErrLogrus.Out = os.Stderr
		formatter.SetLogger(ErrLogrus)
	}
	InitConfigs()

	logrus.Info("Starting root command preparation (`arduino`)")
	if !formatter.IsSupported(GlobalFlags.Format) {
		logrus.WithField("inserted format", GlobalFlags.Format).Warn("Unsupported format, using text as default")
		GlobalFlags.Format = "text"
	}
	formatter.SetFormatter(GlobalFlags.Format)
	logrus.Info("Formatter set")
	if !formatter.IsCurrentFormat("text") {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			logrus.Warn("Calling help on JSON format")
			formatter.PrintErrorMessage("Invalid Call : should show Help, but it is available only in TEXT mode")
			os.Exit(errBadCall)
		})
	}

	if !testing {
		logrus.Info("Initializing viper configuration")
		cobra.OnInitialize(initViper)
	}
}

func arduinoRun(cmd *cobra.Command, args []string) {
	if rootCmdFlags.GenerateDocs {
		logrus.Info("Generating docs")
		errorText := ""
		err := cmd.GenBashCompletionFile("docs/bash_completions/arduino")
		if err != nil {
			logrus.WithError(err).Warn("Error Generating bash autocompletions")
			errorText += fmt.Sprintln(err)
		}
		err = generateManPages(cmd)
		if err != nil {
			logrus.WithError(err).Warn("Error Generating manpages")
			errorText += fmt.Sprintln(err)
		}
		if errorText != "" {
			formatter.PrintErrorMessage(errorText)
		}
	} else {
		logrus.Info("Calling help command")
		cmd.Help()
		os.Exit(errBadCall)
	}
}

// Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	err := ArduinoCmd.Execute()
	if err != nil {
		formatter.PrintError(err, "Bad Exit")
		os.Exit(errGeneric)
	}
}

func executeVersionCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Calling version command on `arduino`")
	versionPrint(versionsToPrint(cmd, true)...)
}

func versionsToPrint(cmd *cobra.Command, isRoot bool) []string {
	verToPrint := make([]string, 0, 10)
	if isRoot {
		verToPrint = append(verToPrint, cmd.Parent().Name())
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

func initViper() {
	logrus.Info("Initiating viper config")

	defHome, err := common.GetDefaultArduinoHomeFolder()
	if err != nil {
		ErrLogrus.WithError(err).Warn("Cannot get default Arduino Home")
	}
	defArduinoData, err := common.GetDefaultArduinoFolder()
	if err != nil {
		logrus.WithError(err).Warn("Cannot get default Arduino folder")
	}

	viper.SetConfigName(".cli-config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	logrus.Info("Reading configuration for viper")
	err = viper.ReadInConfig()
	if err != nil {
		formatter.PrintError(err, "Cannot read configuration file in any of the default folders")
		os.Exit(errNoConfigFile)
	}

	logrus.Info("Setting defaults")
	viper.SetDefault("paths.sketchbook", defHome)
	viper.SetDefault("paths.arduino_data", defArduinoData)
	viper.SetDefault("proxy.type", "auto")
	viper.SetDefault("proxy.hostname", "")
	viper.SetDefault("proxy.username", "")
	viper.SetDefault("proxy.password", "")

	viper.AutomaticEnv()

	logrus.Info("Setting proxy")
	if viper.GetString("proxy.type") == "manual" {
		hostname := viper.GetString("proxy.hostname")
		if hostname == "" {
			ErrLogrus.Error("With manual proxy configuration, hostname is required.")
			formatter.PrintErrorMessage("With manual proxy configuration, hostname is required.")
			os.Exit(errCoreConfig)
		}

		if strings.HasPrefix(hostname, "http") {
			os.Setenv("HTTP_PROXY", hostname)
		}
		if strings.HasPrefix(hostname, "https") {
			os.Setenv("HTTPS_PROXY", hostname)
		}

		username := viper.GetString("proxy.username")
		if username != "" { // put username and pass somewhere

		}
	}
	logrus.Info("Done viper configuration loading")
}

func executeLoginCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino login`")

	userEmpty := arduinoLoginFlags.User == ""
	passwordEmpty := arduinoLoginFlags.Password == ""
	isTextMode := formatter.IsCurrentFormat("text")
	if !isTextMode && (userEmpty || passwordEmpty) {
		formatter.PrintErrorMessage("User and password must be specified outside of text format")
		return
	}

	logrus.Info("Using/Asking credentials")
	if userEmpty {
		fmt.Print("Username: ")
		fmt.Scanln(&arduinoLoginFlags.User)
	}

	if passwordEmpty {
		fmt.Print("Password: ")
		pass, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			formatter.PrintError(err, "Cannot read password, login aborted")
			return
		}
		arduinoLoginFlags.Password = string(pass)
		fmt.Println()
	}

	logrus.Info("Getting ~/.netrc file")

	//save into netrc
	netRCHome, err := homedir.Dir()
	if err != nil {
		formatter.PrintError(err, "Cannot get current home directory")
		os.Exit(errGeneric)
	}

	netRCFile := filepath.Join(netRCHome, ".netrc")
	file, err := os.OpenFile(netRCFile, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		formatter.PrintError(err, "Cannot parse .netrc file")
		return
	}
	defer file.Close()
	NetRC, err := netrc.Parse(file)
	if err != nil {
		formatter.PrintError(err, "Cannot parse .netrc file")
		os.Exit(errGeneric)
	}

	logrus.Info("Trying to login")

	usr := arduinoLoginFlags.User
	pwd := arduinoLoginFlags.Password
	authConf := auth.New()

	token, err := authConf.Token(usr, pwd)
	if err != nil {
		formatter.PrintError(err, "Cannot login")
		os.Exit(errNetwork)
	}

	NetRC.RemoveMachine("arduino.cc")
	NetRC.NewMachine("arduino.cc", usr, token.Access, token.Refresh)
	content, err := NetRC.MarshalText()
	if err != nil {
		formatter.PrintError(err, "Cannot parse new .netrc file")
		os.Exit(errGeneric)
	}

	err = ioutil.WriteFile(netRCFile, content, 0666)
	if err != nil {
		formatter.PrintError(err, "Cannot write new .netrc file")
		os.Exit(errGeneric)
	}

	formatter.PrintResult(`Successfully logged into the system
The session will continue to be refreshed with every call of the CLI and will expire if not used`)
	logrus.Info("Done")
}
