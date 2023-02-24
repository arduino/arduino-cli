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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cli/version"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type showPropertiesMode int

const (
	showPropertiesModeDisabled showPropertiesMode = iota
	showPropertiesModeUnexpanded
	showPropertiesModeExpanded
)

func parseShowPropertiesMode(showProperties string) (showPropertiesMode, error) {
	val, ok := map[string]showPropertiesMode{
		"disabled":   showPropertiesModeDisabled,
		"unexpanded": showPropertiesModeUnexpanded,
		"expanded":   showPropertiesModeExpanded,
	}[showProperties]
	if !ok {
		return showPropertiesModeDisabled, fmt.Errorf(tr("invalid option '%s'.", showProperties))
	}
	return val, nil
}

var (
	fqbnArg                 arguments.Fqbn       // Fully Qualified Board Name, e.g.: arduino:avr:uno.
	profileArg              arguments.Profile    // Profile to use
	showProperties          string               // Show all build preferences used instead of compiling.
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
	dumpProfile             bool                 // Create and print a profile configuration from the build
	// library and libraries sound similar but they're actually different.
	// library expects a path to the root folder of one single library.
	// libraries expects a path to a directory containing multiple libraries, similarly to the <directories.user>/libraries path.
	library                []string // List of paths to libraries root folders. Can be used multiple times for different libraries
	libraries              []string // List of custom libraries dir paths separated by commas. Or can be used multiple times for multiple libraries paths.
	skipLibrariesDiscovery bool
	tr                     = i18n.Tr
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
	compileCommand.Flags().BoolVar(&dumpProfile, "dump-profile", false, tr("Create and print a profile configuration from the build."))
	compileCommand.Flags().StringVar(
		&showProperties,
		"show-properties",
		"disabled",
		tr(`Show build properties instead of compiling. The properties are returned exactly as they are defined. Use "--show-properties=expanded" to replace placeholders with compilation context values.`),
	)
	compileCommand.Flags().Lookup("show-properties").NoOptDefVal = "unexpanded" // default if the flag is present with no value
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
	compileCommand.Flags().BoolVar(&skipLibrariesDiscovery, "skip-libraries-discovery", false, "Skip libraries discovery. This flag is provided only for use in language server and other, very specific, use cases. Do not use for normal compiles")
	compileCommand.Flag("skip-libraries-discovery").Hidden = true
	configuration.Settings.BindPFlag("sketch.always_export_binaries", compileCommand.Flags().Lookup("export-binaries"))

	compileCommand.Flags().MarkDeprecated("build-properties", tr("please use --build-property instead."))

	return compileCommand
}

func runCompileCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino-cli compile`")

	if profileArg.Get() != "" {
		if len(libraries) > 0 {
			feedback.Fatal(tr("You cannot use the %s flag while compiling with a profile.", "--libraries"), feedback.ErrBadArgument)
		}
		if len(library) > 0 {
			feedback.Fatal(tr("You cannot use the %s flag while compiling with a profile.", "--library"), feedback.ErrBadArgument)
		}
	}
	showPropertiesM, err := parseShowPropertiesMode(showProperties)
	if err != nil {
		feedback.Fatal(tr("Error parsing --show-properties flag: %v", err), feedback.ErrGeneric)
	}

	path := ""
	if len(args) > 0 {
		path = args[0]
	}

	sketchPath := arguments.InitSketchPath(path)
	sk, err := arguments.NewSketch(sketchPath)

	if err != nil {
		showPropertiesWithEmptySketchPath := path == "" && showPropertiesM != showPropertiesModeDisabled
		if showPropertiesWithEmptySketchPath {
			// properties were requested and no sketch path was provided
			// let's use an empty sketch struct and hope for the best
			sk = nil
		} else {
			feedback.Fatal(tr("Error opening sketch: %v", err), feedback.ErrGeneric)
		}
	}

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
			feedback.Fatal(tr("Error opening source code overrides data file: %v", err), feedback.ErrGeneric)
		}
		var o struct {
			Overrides map[string]string `json:"overrides"`
		}
		if err := json.Unmarshal(data, &o); err != nil {
			feedback.Fatal(tr("Error: invalid source code overrides data file: %v", err), feedback.ErrGeneric)
		}
		overrides = o.Overrides
	}

	var stdOut, stdErr io.Writer
	var stdIORes func() *feedback.OutputStreamsResult
	if showPropertiesM != showPropertiesModeDisabled {
		stdOut, stdErr, stdIORes = feedback.NewBufferedStreams()
	} else {
		stdOut, stdErr, stdIORes = feedback.OutputStreams()
	}

	compileRequest := &rpc.CompileRequest{
		Instance:                      inst,
		Fqbn:                          fqbn,
		SketchPath:                    sketchPath.String(),
		ShowProperties:                showPropertiesM != showPropertiesModeDisabled,
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
		SkipLibrariesDiscovery:        skipLibrariesDiscovery,
	}
	compileRes, compileError := compile.Compile(context.Background(), compileRequest, stdOut, stdErr, nil)
	if compileError == nil && uploadAfterCompile {
		userFieldRes, err := upload.SupportedUserFields(context.Background(), &rpc.SupportedUserFieldsRequest{
			Instance: inst,
			Fqbn:     fqbn,
			Protocol: port.Protocol,
		})
		if err != nil {
			feedback.Fatal(tr("Error during Upload: %v", err), feedback.ErrGeneric)
		}

		fields := map[string]string{}
		if len(userFieldRes.UserFields) > 0 {
			feedback.Print(tr("Uploading to specified board using %s protocol requires the following info:", port.Protocol))
			if f, err := arguments.AskForUserFields(userFieldRes.UserFields); err != nil {
				feedback.FatalError(err, feedback.ErrBadArgument)
			} else {
				fields = f
			}
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

		if err := upload.Upload(context.Background(), uploadRequest, stdOut, stdErr); err != nil {
			feedback.Fatal(tr("Error during Upload: %v", err), feedback.ErrGeneric)
		}
	}

	profileOut := ""
	if dumpProfile && compileError == nil {
		// Output profile

		libs := ""
		hasVendoredLibs := false
		for _, lib := range compileRes.GetUsedLibraries() {
			if lib.Location != rpc.LibraryLocation_LIBRARY_LOCATION_USER && lib.Location != rpc.LibraryLocation_LIBRARY_LOCATION_UNMANAGED {
				continue
			}
			if lib.GetVersion() == "" {
				hasVendoredLibs = true
				continue
			}
			libs += fmt.Sprintln("      - " + lib.GetName() + " (" + lib.GetVersion() + ")")
		}
		if hasVendoredLibs {
			msg := "\n"
			msg += tr("WARNING: The sketch is compiled using one or more custom libraries.") + "\n"
			msg += tr("Currently, Build Profiles only support libraries available through Arduino Library Manager.")
			feedback.Warning(msg)
		}

		newProfileName := "my_profile_name"
		if split := strings.Split(compileRequest.GetFqbn(), ":"); len(split) > 2 {
			newProfileName = split[2]
		}
		profileOut = fmt.Sprintln("profiles:")
		profileOut += fmt.Sprintln("  " + newProfileName + ":")
		profileOut += fmt.Sprintln("    fqbn: " + compileRequest.GetFqbn())
		profileOut += fmt.Sprintln("    platforms:")
		boardPlatform := compileRes.GetBoardPlatform()
		profileOut += fmt.Sprintln("      - platform: " + boardPlatform.GetId() + " (" + boardPlatform.GetVersion() + ")")
		if url := boardPlatform.GetPackageUrl(); url != "" {
			profileOut += fmt.Sprintln("        platform_index_url: " + url)
		}

		if buildPlatform := compileRes.GetBuildPlatform(); buildPlatform != nil &&
			buildPlatform.Id != boardPlatform.Id &&
			buildPlatform.Version != boardPlatform.Version {
			profileOut += fmt.Sprintln("      - platform: " + buildPlatform.GetId() + " (" + buildPlatform.GetVersion() + ")")
			if url := buildPlatform.GetPackageUrl(); url != "" {
				profileOut += fmt.Sprintln("        platform_index_url: " + url)
			}
		}
		if len(libs) > 0 {
			profileOut += fmt.Sprintln("    libraries:")
			profileOut += fmt.Sprint(libs)
		}
		profileOut += fmt.Sprintln()
	}

	stdIO := stdIORes()
	res := &compileResult{
		CompilerOut:        stdIO.Stdout,
		CompilerErr:        stdIO.Stderr,
		BuilderResult:      compileRes,
		ProfileOut:         profileOut,
		Success:            compileError == nil,
		showPropertiesMode: showPropertiesM,
	}

	if compileError != nil {
		res.Error = tr("Error during build: %v", compileError)

		// Check the error type to give the user better feedback on how
		// to resolve it
		var platformErr *arduino.PlatformNotFoundError
		if errors.As(compileError, &platformErr) {
			split := strings.Split(platformErr.Platform, ":")
			if len(split) < 2 {
				panic(tr("Platform ID is not correct"))
			}

			// FIXME: Here we should not access PackageManager...
			pme, release := commands.GetPackageManagerExplorer(compileRequest)
			platform := pme.FindPlatform(&packagemanager.PlatformReference{
				Package:              split[0],
				PlatformArchitecture: split[1],
			})
			release()

			if profileArg.String() == "" {
				res.Error += fmt.Sprintln()
				if platform != nil {
					suggestion := fmt.Sprintf("`%s core install %s`", version.VersionInfo.Application, platformErr.Platform)
					res.Error += tr("Try running %s", suggestion)
				} else {
					res.Error += tr("Platform %s is not found in any known index\nMaybe you need to add a 3rd party URL?", platformErr.Platform)
				}
			}
		}
		feedback.FatalResult(res, feedback.ErrGeneric)
	}
	if showPropertiesM == showPropertiesModeExpanded {
		expandPropertiesInResult(res)
	}
	feedback.PrintResult(res)
}

func expandPropertiesInResult(res *compileResult) {
	expanded, err := properties.LoadFromSlice(res.BuilderResult.GetBuildProperties())
	if err != nil {
		res.Error = tr(err.Error())
	}
	expandedSlice := make([]string, expanded.Size())
	for i, k := range expanded.Keys() {
		expandedSlice[i] = strings.Join([]string{k, expanded.ExpandPropsInString(expanded.Get(k))}, "=")
	}
	res.BuilderResult.BuildProperties = expandedSlice
}

type compileResult struct {
	CompilerOut   string               `json:"compiler_out"`
	CompilerErr   string               `json:"compiler_err"`
	BuilderResult *rpc.CompileResponse `json:"builder_result"`
	Success       bool                 `json:"success"`
	ProfileOut    string               `json:"profile_out,omitempty"`
	Error         string               `json:"error,omitempty"`

	showPropertiesMode showPropertiesMode
}

func (r *compileResult) Data() interface{} {
	return r
}

func (r *compileResult) String() string {
	if r.showPropertiesMode != showPropertiesModeDisabled {
		return strings.Join(r.BuilderResult.GetBuildProperties(), fmt.Sprintln())
	}

	titleColor := color.New(color.FgHiGreen)
	nameColor := color.New(color.FgHiYellow)
	pathColor := color.New(color.FgHiBlack)
	build := r.BuilderResult

	res := ""
	if r.CompilerOut != "" || r.CompilerErr != "" {
		res += fmt.Sprintln()
	}
	if len(build.GetUsedLibraries()) > 0 {
		libraries := table.New()
		libraries.SetHeader(
			table.NewCell(tr("Used library"), titleColor),
			table.NewCell(tr("Version"), titleColor),
			table.NewCell(tr("Path"), pathColor))
		for _, l := range build.GetUsedLibraries() {
			libraries.AddRow(
				table.NewCell(l.GetName(), nameColor),
				l.GetVersion(),
				table.NewCell(l.GetInstallDir(), pathColor))
		}
		res += fmt.Sprintln(libraries.Render())
	}

	if boardPlatform := build.GetBoardPlatform(); boardPlatform != nil {
		platforms := table.New()
		platforms.SetHeader(
			table.NewCell(tr("Used platform"), titleColor),
			table.NewCell(tr("Version"), titleColor),
			table.NewCell(tr("Path"), pathColor))
		platforms.AddRow(
			table.NewCell(boardPlatform.GetId(), nameColor),
			boardPlatform.GetVersion(),
			table.NewCell(boardPlatform.GetInstallDir(), pathColor))
		if buildPlatform := build.GetBuildPlatform(); buildPlatform != nil &&
			buildPlatform.Id != boardPlatform.Id &&
			buildPlatform.Version != boardPlatform.Version {
			platforms.AddRow(
				table.NewCell(buildPlatform.GetId(), nameColor),
				buildPlatform.GetVersion(),
				table.NewCell(buildPlatform.GetInstallDir(), pathColor))
		}
		res += fmt.Sprintln(platforms.Render())
	}
	if r.ProfileOut != "" {
		res += fmt.Sprintln(r.ProfileOut)
	}
	return strings.TrimRight(res, fmt.Sprintln())
}

func (r *compileResult) ErrorString() string {
	return r.Error
}
