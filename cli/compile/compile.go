// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package compile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/sirupsen/logrus"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/commands/upload"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

var (
	fqbnArg                 arguments.Fqbn       // Fully Qualified Board Name, e.g.: arduino:avr:uno.
	profileArg              arguments.Profile    // Profile to use
	showProperties          bool                 // Show all build preferences used instead of compiling.
	preprocess              bool                 // Print preprocessed code to stdout.
	buildCachePath          string               // Builds of 'core.a' are saved into this path to be cached and reused.
	buildPath               string               // Path where to save compiled files.
	buildProperties         []string             // List of custom build properties separated by commas. Or can be used multiple times for multiple properties.
	keysKeychain            string               // The path of the dir where to search for the custom keys to sign and encrypt a binary. Used only by the platforms that supports it
	signKey                 string               // The name of the custom signing key to use to sign a binary during the compile process. Used only by the platforms that supports it
	encryptKey              string               // The name of the custom encryption key to use to encrypt a binary during the compile process. Used only by the platforms that supports it
	warnings                string               // Used to tell gcc which warning level to use.
	verbose                 bool                 // Turns on verbose mode.
	quiet                   bool                 // Suppresses almost every output.
	vidPid                  string               // VID/PID specific build properties.
	uploadAfterCompile      bool                 // Upload the binary after the compilation.
	portArgs                arguments.Port       // Upload port, e.g.: COM10 or /dev/ttyACM0.
	verify                  bool                 // Upload, verify uploaded binary after the upload.
	exportDir               string               // The compiled binary is written to this file
	optimizeForDebug        bool                 // Optimize compile output for debug, not for release
	programmer              arguments.Programmer // Use the specified programmer to upload
	clean                   bool                 // Cleanup the build folder and do not use any cached build
	compilationDatabaseOnly bool                 // Only create compilation database without actually compiling
	sourceOverrides         string               // Path to a .json file that contains a set of replacements of the sketch source code.
	// library and libraries sound similar but they're actually different.
	// library expects a path to the root folder of one single library.
	// libraries expects a path to a directory containing multiple libraries, similarly to the <directories.user>/libraries path.
	library   []string // List of paths to libraries root folders. Can be used multiple times for different libraries
	libraries []string // List of custom libraries dir paths separated by commas. Or can be used multiple times for multiple libraries paths.
	tr        = i18n.Tr
)

// NewCommand created a new `compile` command
func NewCommand() *cobra.Command {
	compileCommand := &cobra.Command{
		Use:   "compile",
		Short: tr("Compiles Arduino sketches."),
		Long:  tr("Compiles Arduino sketches."),
		Example: "" +
			"  " + os.Args[0] + " compile -b arduino:avr:uno /home/user/Arduino/MySketch\n" +
			"  " + os.Args[0] + ` compile -b arduino:avr:uno --build-property "build.extra_flags=\"-DMY_DEFINE=\"hello world\"\"" /home/user/Arduino/MySketch` + "\n" +
			"  " + os.Args[0] + ` compile -b arduino:avr:uno --build-property "build.extra_flags=-DPIN=2 \"-DMY_DEFINE=\"hello world\"\"" /home/user/Arduino/MySketch` + "\n" +
			"  " + os.Args[0] + ` compile -b arduino:avr:uno --build-property build.extra_flags=-DPIN=2 --build-property "compiler.cpp.extra_flags=\"-DSSID=\"hello world\"\"" /home/user/Arduino/MySketch` + "\n",
		Args: cobra.MaximumNArgs(1),
		Run:  runCompileCommand,
	}

	fqbnArg.AddToCommand(compileCommand)
	profileArg.AddToCommand(compileCommand)
	compileCommand.Flags().BoolVar(&showProperties, "show-properties", false, tr("Show all build properties used instead of compiling."))
	compileCommand.Flags().BoolVar(&preprocess, "preprocess", false, tr("Print preprocessed code to stdout instead of compiling."))
	compileCommand.Flags().StringVar(&buildCachePath, "build-cache-path", "", tr("Builds of 'core.a' are saved into this path to be cached and reused."))
	compileCommand.Flags().StringVarP(&exportDir, "output-dir", "", "", tr("Save build artifacts in this directory."))
	compileCommand.Flags().StringVar(&buildPath, "build-path", "",
		tr("Path where to save compiled files. If omitted, a directory will be created in the default temporary path of your OS."))
	compileCommand.Flags().StringSliceVar(&buildProperties, "build-properties", []string{},
		tr("List of custom build properties separated by commas. Or can be used multiple times for multiple properties."))
	compileCommand.Flags().StringArrayVar(&buildProperties, "build-property", []string{},
		tr("Override a build property with a custom value. Can be used multiple times for multiple properties."))
	compileCommand.Flags().StringVar(&keysKeychain, "keys-keychain", "",
		tr("The path of the dir to search for the custom keys to sign and encrypt a binary. Used only by the platforms that support it."))
	compileCommand.Flags().StringVar(&signKey, "sign-key", "",
		tr("The name of the custom signing key to use to sign a binary during the compile process. Used only by the platforms that support it."))
	compileCommand.Flags().StringVar(&encryptKey, "encrypt-key", "",
		tr("The name of the custom encryption key to use to encrypt a binary during the compile process. Used only by the platforms that support it."))
	compileCommand.Flags().StringVar(&warnings, "warnings", "none",
		tr(`Optional, can be: %s. Used to tell gcc which warning level to use (-W flag).`, "none, default, more, all"))
	compileCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, tr("Optional, turns on verbose mode."))
	compileCommand.Flags().BoolVar(&quiet, "quiet", false, tr("Optional, suppresses almost every output."))
	compileCommand.Flags().BoolVarP(&uploadAfterCompile, "upload", "u", false, tr("Upload the binary after the compilation."))
	portArgs.AddToCommand(compileCommand)
	compileCommand.Flags().BoolVarP(&verify, "verify", "t", false, tr("Verify uploaded binary after the upload."))
	compileCommand.Flags().StringVar(&vidPid, "vid-pid", "", tr("When specified, VID/PID specific build properties are used, if board supports them."))
	compileCommand.Flags().StringSliceVar(&library, "library", []string{},
		tr("Path to a single library’s root folder. Can be used multiple times or entries can be comma separated."))
	compileCommand.Flags().StringSliceVar(&libraries, "libraries", []string{},
		tr("Path to a collection of libraries. Can be used multiple times or entries can be comma separated."))
	compileCommand.Flags().BoolVar(&optimizeForDebug, "optimize-for-debug", false, tr("Optional, optimize compile output for debugging, rather than for release."))
	programmer.AddToCommand(compileCommand)
	compileCommand.Flags().BoolVar(&compilationDatabaseOnly, "only-compilation-database", false, tr("Just produce the compilation database, without actually compiling. All build commands are skipped except pre* hooks."))
	compileCommand.Flags().BoolVar(&clean, "clean", false, tr("Optional, cleanup the build folder and do not use any cached build."))
	// We must use the following syntax for this flag since it's also bound to settings.
	// This must be done because the value is set when the binding is accessed from viper. Accessing from cobra would only
	// read the value if the flag is set explicitly by the user.
	compileCommand.Flags().BoolP("export-binaries", "e", false, tr("If set built binaries will be exported to the sketch folder."))
	compileCommand.Flags().StringVar(&sourceOverrides, "source-override", "", tr("Optional. Path to a .json file that contains a set of replacements of the sketch source code."))
	compileCommand.Flag("source-override").Hidden = true

	configuration.Settings.BindPFlag("sketch.always_export_binaries", compileCommand.Flags().Lookup("export-binaries"))

	compileCommand.Flags().MarkDeprecated("build-properties", tr("please use --build-property instead."))

	return compileCommand
}

func runCompileCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino-cli compile`")

	if profileArg.Get() != "" {
		if len(libraries) > 0 {
			feedback.Errorf(tr("You cannot use the %s flag while compiling with a profile.", "--libraries"))
			os.Exit(errorcodes.ErrBadArgument)
		}
		if len(library) > 0 {
			feedback.Errorf(tr("You cannot use the %s flag while compiling with a profile.", "--library"))
			os.Exit(errorcodes.ErrBadArgument)
		}
	}

	path := ""
	if len(args) > 0 {
		path = args[0]
	}

	sketchPath := arguments.InitSketchPath(path)
	sk := arguments.NewSketch(sketchPath)

	inst, profile := instance.CreateAndInitWithProfile(profileArg.Get(), sketchPath)
	if fqbnArg.String() == "" {
		fqbnArg.Set(profile.GetFqbn())
	}

	fqbn, port := arguments.CalculateFQBNAndPort(&portArgs, &fqbnArg, inst, sk)

	if keysKeychain != "" || signKey != "" || encryptKey != "" {
		arguments.CheckFlagsMandatory(cmd, "keys-keychain", "sign-key", "encrypt-key")
	}

	var overrides map[string]string
	if sourceOverrides != "" {
		data, err := paths.New(sourceOverrides).ReadFile()
		if err != nil {
			feedback.Errorf(tr("Error opening source code overrides data file: %v"), err)
			os.Exit(errorcodes.ErrGeneric)
		}
		var o struct {
			Overrides map[string]string `json:"overrides"`
		}
		if err := json.Unmarshal(data, &o); err != nil {
			feedback.Errorf(tr("Error: invalid source code overrides data file: %v"), err)
			os.Exit(errorcodes.ErrGeneric)
		}
		overrides = o.Overrides
	}

	compileRequest := &rpc.CompileRequest{
		Instance:                      inst,
		Fqbn:                          fqbn,
		SketchPath:                    sketchPath.String(),
		ShowProperties:                showProperties,
		Preprocess:                    preprocess,
		BuildCachePath:                buildCachePath,
		BuildPath:                     buildPath,
		BuildProperties:               buildProperties,
		Warnings:                      warnings,
		Verbose:                       verbose,
		Quiet:                         quiet,
		VidPid:                        vidPid,
		ExportDir:                     exportDir,
		Libraries:                     libraries,
		OptimizeForDebug:              optimizeForDebug,
		Clean:                         clean,
		CreateCompilationDatabaseOnly: compilationDatabaseOnly,
		SourceOverride:                overrides,
		Library:                       library,
		KeysKeychain:                  keysKeychain,
		SignKey:                       signKey,
		EncryptKey:                    encryptKey,
	}
	compileStdOut := new(bytes.Buffer)
	compileStdErr := new(bytes.Buffer)
	verboseCompile := configuration.Settings.GetString("logging.level") == "debug"
	var compileRes *rpc.CompileResponse
	var compileError error
	if output.OutputFormat == "json" {
		compileRes, compileError = compile.Compile(context.Background(), compileRequest, compileStdOut, compileStdErr, nil, verboseCompile)
	} else {
		compileRes, compileError = compile.Compile(context.Background(), compileRequest, os.Stdout, os.Stderr, nil, verboseCompile)
	}

	if compileError == nil && uploadAfterCompile {
		userFieldRes, err := upload.SupportedUserFields(context.Background(), &rpc.SupportedUserFieldsRequest{
			Instance: inst,
			Fqbn:     fqbn,
			Protocol: port.Protocol,
		})
		if err != nil {
			feedback.Errorf(tr("Error during Upload: %v", err))
			os.Exit(errorcodes.ErrGeneric)
		}

		fields := map[string]string{}
		if len(userFieldRes.UserFields) > 0 {
			feedback.Print(tr("Uploading to specified board using %s protocol requires the following info:", port.Protocol))
			fields = arguments.AskForUserFields(userFieldRes.UserFields)
		}

		uploadRequest := &rpc.UploadRequest{
			Instance:   inst,
			Fqbn:       fqbn,
			SketchPath: sketchPath.String(),
			Port:       port,
			Verbose:    verbose,
			Verify:     verify,
			ImportDir:  buildPath,
			Programmer: programmer.String(),
			UserFields: fields,
		}

		var uploadError error
		if output.OutputFormat == "json" {
			// TODO: do not print upload output in json mode
			uploadStdOut := new(bytes.Buffer)
			uploadStdErr := new(bytes.Buffer)
			_, uploadError = upload.Upload(context.Background(), uploadRequest, uploadStdOut, uploadStdErr)
		} else {
			_, uploadError = upload.Upload(context.Background(), uploadRequest, os.Stdout, os.Stderr)
		}
		if uploadError != nil {
			feedback.Errorf(tr("Error during Upload: %v"), uploadError)
			os.Exit(errorcodes.ErrGeneric)
		}
	}

	feedback.PrintResult(&compileResult{
		CompileOut:    compileStdOut.String(),
		CompileErr:    compileStdErr.String(),
		BuilderResult: compileRes,
		Success:       compileError == nil,
	})
	if compileError != nil {
		feedback.Errorf(tr("Error during build: %v"), compileError)

		// Check the error type to give the user better feedback on how
		// to resolve it
		var platformErr *arduino.PlatformNotFoundError
		if errors.As(compileError, &platformErr) {
			split := strings.Split(platformErr.Platform, ":")
			if len(split) < 2 {
				panic(tr("Platform ID is not correct"))
			}

			pm := commands.GetPackageManager(inst.GetId())
			platform := pm.FindPlatform(&packagemanager.PlatformReference{
				Package:              split[0],
				PlatformArchitecture: split[1],
			})

			if profileArg.String() == "" {
				if platform != nil {
					feedback.Errorf(tr("Try running %s", fmt.Sprintf("`%s core install %s`", globals.VersionInfo.Application, platformErr.Platform)))
				} else {
					feedback.Errorf(tr("Platform %s is not found in any known index\nMaybe you need to add a 3rd party URL?", platformErr.Platform))
				}
			}
		}
		os.Exit(errorcodes.ErrGeneric)
	}
}

type compileResult struct {
	CompileOut    string               `json:"compiler_out"`
	CompileErr    string               `json:"compiler_err"`
	BuilderResult *rpc.CompileResponse `json:"builder_result"`
	Success       bool                 `json:"success"`
}

func (r *compileResult) Data() interface{} {
	return r
}

func (r *compileResult) String() string {
	// The output is already printed via os.Stdout/os.Stdin
	return ""
}
