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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/feedback/table"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/internal/version"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbnArg                 arguments.Fqbn           // Fully Qualified Board Name, e.g.: arduino:avr:uno.
	profileArg              arguments.Profile        // Profile to use
	showPropertiesArg       arguments.ShowProperties // Show all build preferences used instead of compiling.
	preprocess              bool                     // Print preprocessed code to stdout.
	buildCachePath          string                   // Builds of 'core.a' are saved into this path to be cached and reused.
	buildPath               string                   // Path where to save compiled files.
	buildProperties         []string                 // List of custom build properties separated by commas. Or can be used multiple times for multiple properties.
	keysKeychain            string                   // The path of the dir where to search for the custom keys to sign and encrypt a binary. Used only by the platforms that supports it
	signKey                 string                   // The name of the custom signing key to use to sign a binary during the compile process. Used only by the platforms that supports it
	encryptKey              string                   // The name of the custom encryption key to use to encrypt a binary during the compile process. Used only by the platforms that supports it
	warnings                string                   // Used to tell gcc which warning level to use.
	verbose                 bool                     // Turns on verbose mode.
	quiet                   bool                     // Suppresses almost every output.
	uploadAfterCompile      bool                     // Upload the binary after the compilation.
	portArgs                arguments.Port           // Upload port, e.g.: COM10 or /dev/ttyACM0.
	verify                  bool                     // Upload, verify uploaded binary after the upload.
	exportBinaries          bool                     // If set built binaries will be exported to the sketch folder
	exportDir               string                   // The compiled binary is written to this file
	optimizeForDebug        bool                     // Optimize compile output for debug, not for release
	programmer              arguments.Programmer     // Use the specified programmer to upload
	clean                   bool                     // Cleanup the build folder and do not use any cached build
	compilationDatabaseOnly bool                     // Only create compilation database without actually compiling
	sourceOverrides         string                   // Path to a .json file that contains a set of replacements of the sketch source code.
	dumpProfile             bool                     // Create and print a profile configuration from the build
	jobs                    int32                    // Max number of parallel jobs
	// library and libraries sound similar but they're actually different.
	// library expects a path to the root folder of one single library.
	// libraries expects a path to a directory containing multiple libraries, similarly to the <directories.user>/libraries path.
	library                []string // List of paths to libraries root folders. Can be used multiple times for different libraries
	libraries              []string // List of custom libraries dir paths separated by commas. Or can be used multiple times for multiple libraries paths.
	skipLibrariesDiscovery bool
)

// NewCommand created a new `compile` command
func NewCommand(srv rpc.ArduinoCoreServiceServer, settings *rpc.Configuration) *cobra.Command {
	compileCommand := &cobra.Command{
		Use:   "compile",
		Short: i18n.Tr("Compiles Arduino sketches."),
		Long:  i18n.Tr("Compiles Arduino sketches."),
		Example: "" +
			"  " + os.Args[0] + " compile -b arduino:avr:uno /home/user/Arduino/MySketch\n" +
			"  " + os.Args[0] + ` compile -b arduino:avr:uno --build-property "build.extra_flags=\"-DMY_DEFINE=\"hello world\"\"" /home/user/Arduino/MySketch` + "\n" +
			"  " + os.Args[0] + ` compile -b arduino:avr:uno --build-property "build.extra_flags=-DPIN=2 \"-DMY_DEFINE=\"hello world\"\"" /home/user/Arduino/MySketch` + "\n" +
			"  " + os.Args[0] + ` compile -b arduino:avr:uno --build-property build.extra_flags=-DPIN=2 --build-property "compiler.cpp.extra_flags=\"-DSSID=\"hello world\"\"" /home/user/Arduino/MySketch` + "\n",
		Args: cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			arguments.CheckFlagsConflicts(cmd, "quiet", "verbose")
		},
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flag("build-cache-path").Changed {
				feedback.Warning(i18n.Tr("The flag --build-cache-path has been deprecated. Please use just --build-path alone or configure the build cache path in the Arduino CLI settings."))
			}
			runCompileCommand(cmd, args, srv)
		},
	}

	fqbnArg.AddToCommand(compileCommand, srv)
	profileArg.AddToCommand(compileCommand, srv)
	compileCommand.Flags().BoolVar(&dumpProfile, "dump-profile", false, i18n.Tr("Create and print a profile configuration from the build."))
	showPropertiesArg.AddToCommand(compileCommand)
	compileCommand.Flags().BoolVar(&preprocess, "preprocess", false, i18n.Tr("Print preprocessed code to stdout instead of compiling."))
	compileCommand.Flags().StringVar(&buildCachePath, "build-cache-path", "", i18n.Tr("Builds of cores and sketches are saved into this path to be cached and reused."))
	compileCommand.Flag("build-cache-path").Hidden = true // deprecated
	compileCommand.Flags().StringVar(&exportDir, "output-dir", "", i18n.Tr("Save build artifacts in this directory."))
	compileCommand.Flags().StringVar(&buildPath, "build-path", "",
		i18n.Tr("Path where to save compiled files. If omitted, a directory will be created in the default temporary path of your OS."))
	compileCommand.Flags().StringSliceVar(&buildProperties, "build-properties", []string{},
		i18n.Tr("List of custom build properties separated by commas. Or can be used multiple times for multiple properties."))
	compileCommand.Flags().StringArrayVar(&buildProperties, "build-property", []string{},
		i18n.Tr("Override a build property with a custom value. Can be used multiple times for multiple properties."))
	compileCommand.Flags().StringVar(&keysKeychain, "keys-keychain", "",
		i18n.Tr("The path of the dir to search for the custom keys to sign and encrypt a binary. Used only by the platforms that support it."))
	compileCommand.Flags().StringVar(&signKey, "sign-key", "",
		i18n.Tr("The name of the custom signing key to use to sign a binary during the compile process. Used only by the platforms that support it."))
	compileCommand.Flags().StringVar(&encryptKey, "encrypt-key", "",
		i18n.Tr("The name of the custom encryption key to use to encrypt a binary during the compile process. Used only by the platforms that support it."))
	compileCommand.Flags().StringVar(&warnings, "warnings", "none",
		i18n.Tr(`Optional, can be: %s. Used to tell gcc which warning level to use (-W flag).`, "none, default, more, all"))
	compileCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, i18n.Tr("Optional, turns on verbose mode."))
	compileCommand.Flags().BoolVarP(&quiet, "quiet", "q", false, i18n.Tr("Optional, suppresses almost every output."))
	compileCommand.Flags().BoolVarP(&uploadAfterCompile, "upload", "u", false, i18n.Tr("Upload the binary after the compilation."))
	portArgs.AddToCommand(compileCommand, srv)
	compileCommand.Flags().BoolVarP(&verify, "verify", "t", false, i18n.Tr("Verify uploaded binary after the upload."))
	compileCommand.Flags().StringSliceVar(&library, "library", []string{},
		i18n.Tr("Path to a single library’s root folder. Can be used multiple times or entries can be comma separated."))
	compileCommand.Flags().StringSliceVar(&libraries, "libraries", []string{},
		i18n.Tr("Path to a collection of libraries. Can be used multiple times or entries can be comma separated."))
	compileCommand.Flags().BoolVar(&optimizeForDebug, "optimize-for-debug", false, i18n.Tr("Optional, optimize compile output for debugging, rather than for release."))
	programmer.AddToCommand(compileCommand, srv)
	compileCommand.Flags().BoolVar(&compilationDatabaseOnly, "only-compilation-database", false, i18n.Tr("Just produce the compilation database, without actually compiling. All build commands are skipped except pre* hooks."))
	compileCommand.Flags().BoolVar(&clean, "clean", false, i18n.Tr("Optional, cleanup the build folder and do not use any cached build."))
	compileCommand.Flags().BoolVarP(&exportBinaries, "export-binaries", "e", settings.GetSketch().GetAlwaysExportBinaries(),
		i18n.Tr("If set built binaries will be exported to the sketch folder."))
	compileCommand.Flags().StringVar(&sourceOverrides, "source-override", "", i18n.Tr("Optional. Path to a .json file that contains a set of replacements of the sketch source code."))
	compileCommand.Flag("source-override").Hidden = true
	compileCommand.Flags().BoolVar(&skipLibrariesDiscovery, "skip-libraries-discovery", false, "Skip libraries discovery. This flag is provided only for use in language server and other, very specific, use cases. Do not use for normal compiles")
	compileCommand.Flag("skip-libraries-discovery").Hidden = true
	compileCommand.Flags().Int32VarP(&jobs, "jobs", "j", 0, i18n.Tr("Max number of parallel compiles. If set to 0 the number of available CPUs cores will be used."))

	compileCommand.Flags().MarkDeprecated("build-properties", i18n.Tr("please use --build-property instead."))

	return compileCommand
}

func runCompileCommand(cmd *cobra.Command, args []string, srv rpc.ArduinoCoreServiceServer) {
	logrus.Info("Executing `arduino-cli compile`")
	ctx := cmd.Context()

	if profileArg.Get() != "" {
		if len(libraries) > 0 {
			feedback.Fatal(i18n.Tr("You cannot use the %s flag while compiling with a profile.", "--libraries"), feedback.ErrBadArgument)
		}
		if len(library) > 0 {
			feedback.Fatal(i18n.Tr("You cannot use the %s flag while compiling with a profile.", "--library"), feedback.ErrBadArgument)
		}
	}

	path := ""
	if len(args) > 0 {
		path = args[0]
	}

	sketchPath := arguments.InitSketchPath(path)
	resp, err := srv.LoadSketch(ctx, &rpc.LoadSketchRequest{SketchPath: sketchPath.String()})
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}
	sk := resp.GetSketch()
	feedback.WarnAboutDeprecatedFiles(sk)

	var inst *rpc.Instance
	var profile *rpc.SketchProfile

	if profileArg.Get() == "" {
		inst, profile = instance.CreateAndInitWithProfile(ctx, srv, sk.GetDefaultProfile().GetName(), sketchPath)
	} else {
		inst, profile = instance.CreateAndInitWithProfile(ctx, srv, profileArg.Get(), sketchPath)
	}

	if fqbnArg.String() == "" {
		fqbnArg.Set(profile.GetFqbn())
	}

	fqbn, port := arguments.CalculateFQBNAndPort(ctx, &portArgs, &fqbnArg, inst, srv, sk.GetDefaultFqbn(), sk.GetDefaultPort(), sk.GetDefaultProtocol(), profile)

	if keysKeychain != "" || signKey != "" || encryptKey != "" {
		arguments.CheckFlagsMandatory(cmd, "keys-keychain", "sign-key", "encrypt-key")
	}

	var overrides map[string]string
	if sourceOverrides != "" {
		data, err := paths.New(sourceOverrides).ReadFile()
		if err != nil {
			feedback.Fatal(i18n.Tr("Error opening source code overrides data file: %v", err), feedback.ErrGeneric)
		}
		var o struct {
			Overrides map[string]string `json:"overrides"`
		}
		if err := json.Unmarshal(data, &o); err != nil {
			feedback.Fatal(i18n.Tr("Error: invalid source code overrides data file: %v", err), feedback.ErrGeneric)
		}
		overrides = o.Overrides
	}

	showProperties, err := showPropertiesArg.Get()
	if err != nil {
		feedback.Fatal(i18n.Tr("Error parsing --show-properties flag: %v", err), feedback.ErrGeneric)
	}

	var stdOut, stdErr io.Writer
	var stdIORes func() *feedback.OutputStreamsResult
	if showProperties != arguments.ShowPropertiesDisabled {
		stdOut, stdErr, stdIORes = feedback.NewBufferedStreams()
	} else {
		stdOut, stdErr, stdIORes = feedback.OutputStreams()
	}

	var libraryAbs []string
	for _, libPath := range paths.NewPathList(library...) {
		if libPath, err = libPath.Abs(); err != nil {
			feedback.Fatal(i18n.Tr("Error converting path to absolute: %v", err), feedback.ErrGeneric)
		}
		libraryAbs = append(libraryAbs, libPath.String())
	}

	compileRequest := &rpc.CompileRequest{
		Instance:                      inst,
		Fqbn:                          fqbn,
		SketchPath:                    sketchPath.String(),
		ShowProperties:                showProperties != arguments.ShowPropertiesDisabled,
		Preprocess:                    preprocess,
		BuildCachePath:                buildCachePath,
		BuildPath:                     buildPath,
		BuildProperties:               buildProperties,
		Warnings:                      warnings,
		Verbose:                       verbose,
		Quiet:                         quiet,
		ExportBinaries:                &exportBinaries,
		ExportDir:                     exportDir,
		Libraries:                     libraries,
		OptimizeForDebug:              optimizeForDebug,
		Clean:                         clean,
		CreateCompilationDatabaseOnly: compilationDatabaseOnly,
		SourceOverride:                overrides,
		Library:                       libraryAbs,
		KeysKeychain:                  keysKeychain,
		SignKey:                       signKey,
		EncryptKey:                    encryptKey,
		SkipLibrariesDiscovery:        skipLibrariesDiscovery,
		DoNotExpandBuildProperties:    showProperties == arguments.ShowPropertiesUnexpanded,
		Jobs:                          jobs,
	}
	server, builderResCB := commands.CompilerServerToStreams(ctx, stdOut, stdErr, nil)
	compileError := srv.Compile(compileRequest, server)
	builderRes := builderResCB()

	var uploadRes *rpc.UploadResult
	if compileError == nil && uploadAfterCompile {
		userFieldRes, err := srv.SupportedUserFields(ctx, &rpc.SupportedUserFieldsRequest{
			Instance: inst,
			Fqbn:     fqbn,
			Protocol: port.GetProtocol(),
		})
		if err != nil {
			feedback.Fatal(i18n.Tr("Error during Upload: %v", err), feedback.ErrGeneric)
		}

		fields := map[string]string{}
		if len(userFieldRes.GetUserFields()) > 0 {
			feedback.Print(i18n.Tr("Uploading to specified board using %s protocol requires the following info:", port.GetProtocol()))
			if f, err := arguments.AskForUserFields(userFieldRes.GetUserFields()); err != nil {
				feedback.FatalError(err, feedback.ErrBadArgument)
			} else {
				fields = f
			}
		}

		prog := profile.GetProgrammer()
		if prog == "" || programmer.GetProgrammer() != "" {
			prog = programmer.String(ctx, inst, srv, fqbn)
		}
		if prog == "" {
			prog = sk.GetDefaultProgrammer()
		}

		uploadRequest := &rpc.UploadRequest{
			Instance:   inst,
			Fqbn:       fqbn,
			SketchPath: sketchPath.String(),
			Port:       port,
			Verbose:    verbose,
			Verify:     verify,
			ImportDir:  buildPath,
			Programmer: prog,
			UserFields: fields,
		}

		stream, streamRes := commands.UploadToServerStreams(ctx, stdOut, stdErr)
		if err := srv.Upload(uploadRequest, stream); err != nil {
			errcode := feedback.ErrGeneric
			if errors.Is(err, &cmderrors.ProgrammerRequiredForUploadError{}) {
				errcode = feedback.ErrMissingProgrammer
			}
			if errors.Is(err, &cmderrors.MissingProgrammerError{}) {
				errcode = feedback.ErrMissingProgrammer
			}
			feedback.Fatal(i18n.Tr("Error during Upload: %v", err), errcode)
		} else {
			uploadRes = streamRes()
		}
	}

	profileOut := ""
	if dumpProfile && compileError == nil {
		// Output profile

		libs := ""
		hasVendoredLibs := false
		for _, lib := range builderRes.GetUsedLibraries() {
			if lib.GetLocation() != rpc.LibraryLocation_LIBRARY_LOCATION_USER && lib.GetLocation() != rpc.LibraryLocation_LIBRARY_LOCATION_UNMANAGED {
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
			msg += i18n.Tr("WARNING: The sketch is compiled using one or more custom libraries.") + "\n"
			msg += i18n.Tr("Currently, Build Profiles only support libraries available through Arduino Library Manager.")
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
		boardPlatform := builderRes.GetBoardPlatform()
		profileOut += fmt.Sprintln("      - platform: " + boardPlatform.GetId() + " (" + boardPlatform.GetVersion() + ")")
		if url := boardPlatform.GetPackageUrl(); url != "" {
			profileOut += fmt.Sprintln("        platform_index_url: " + url)
		}

		if buildPlatform := builderRes.GetBuildPlatform(); buildPlatform != nil &&
			buildPlatform.GetId() != boardPlatform.GetId() &&
			buildPlatform.GetVersion() != boardPlatform.GetVersion() {
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
	successful := (compileError == nil)
	res := &compileResult{
		CompilerOut:   stdIO.Stdout,
		CompilerErr:   stdIO.Stderr,
		BuilderResult: result.NewBuilderResult(builderRes),
		UploadResult: updatedUploadPortResult{
			UpdatedUploadPort: result.NewPort(uploadRes.GetUpdatedUploadPort()),
		},
		ProfileOut:         profileOut,
		Success:            successful,
		showPropertiesMode: showProperties,
		hideStats:          preprocess || quiet || (!verbose && successful),
	}

	if compileError != nil {
		res.Error = i18n.Tr("Error during build: %v", compileError)

		// Check the error type to give the user better feedback on how
		// to resolve it
		var platformErr *cmderrors.PlatformNotFoundError
		if errors.As(compileError, &platformErr) {
			split := strings.Split(platformErr.Platform, ":")
			if len(split) < 2 {
				panic(i18n.Tr("Platform ID is not correct"))
			}

			if profileArg.String() == "" {
				res.Error += fmt.Sprintln()

				if platform, err := srv.PlatformSearch(ctx, &rpc.PlatformSearchRequest{
					Instance:   inst,
					SearchArgs: platformErr.Platform,
				}); err != nil {
					res.Error += err.Error()
				} else if len(platform.GetSearchOutput()) > 0 {
					suggestion := fmt.Sprintf("`%s core install %s`", version.VersionInfo.Application, platformErr.Platform)
					res.Error += i18n.Tr("Try running %s", suggestion)
				} else {
					res.Error += i18n.Tr("Platform %s is not found in any known index\nMaybe you need to add a 3rd party URL?", platformErr.Platform)
				}
			}
		}
		feedback.FatalResult(res, feedback.ErrGeneric)
	}
	feedback.PrintResult(res)
}

type updatedUploadPortResult struct {
	UpdatedUploadPort *result.Port `json:"updated_upload_port,omitempty"`
}

type compileResult struct {
	CompilerOut        string                  `json:"compiler_out"`
	CompilerErr        string                  `json:"compiler_err"`
	BuilderResult      *result.BuilderResult   `json:"builder_result"`
	UploadResult       updatedUploadPortResult `json:"upload_result"`
	Success            bool                    `json:"success"`
	ProfileOut         string                  `json:"profile_out,omitempty"`
	Error              string                  `json:"error,omitempty"`
	showPropertiesMode arguments.ShowPropertiesMode
	hideStats          bool
}

func (r *compileResult) Data() interface{} {
	return r
}

func (r *compileResult) String() string {
	if r.BuilderResult != nil && r.showPropertiesMode != arguments.ShowPropertiesDisabled {
		return strings.Join(r.BuilderResult.BuildProperties, fmt.Sprintln())
	}

	if r.hideStats {
		return ""
	}

	titleColor := color.New(color.FgHiGreen)
	nameColor := color.New(color.FgHiYellow)
	pathColor := color.New(color.FgHiBlack)
	build := r.BuilderResult

	res := ""
	if r.CompilerOut != "" || r.CompilerErr != "" {
		res += fmt.Sprintln()
	}
	if build != nil && len(build.UsedLibraries) > 0 {
		libraries := table.New()
		libraries.SetHeader(
			table.NewCell(i18n.Tr("Used library"), titleColor),
			table.NewCell(i18n.Tr("Version"), titleColor),
			table.NewCell(i18n.Tr("Path"), pathColor))
		for _, l := range build.UsedLibraries {
			libraries.AddRow(
				table.NewCell(l.Name, nameColor),
				l.Version,
				table.NewCell(l.InstallDir, pathColor))
		}
		res += fmt.Sprintln(libraries.Render())
	}
	if build != nil && build.BoardPlatform != nil {
		boardPlatform := build.BoardPlatform
		platforms := table.New()
		platforms.SetHeader(
			table.NewCell(i18n.Tr("Used platform"), titleColor),
			table.NewCell(i18n.Tr("Version"), titleColor),
			table.NewCell(i18n.Tr("Path"), pathColor))
		platforms.AddRow(
			table.NewCell(boardPlatform.Id, nameColor),
			boardPlatform.Version,
			table.NewCell(boardPlatform.InstallDir, pathColor))
		if buildPlatform := build.BuildPlatform; buildPlatform != nil &&
			buildPlatform.Id != boardPlatform.Id &&
			buildPlatform.Version != boardPlatform.Version {
			platforms.AddRow(
				table.NewCell(buildPlatform.Id, nameColor),
				buildPlatform.Version,
				table.NewCell(buildPlatform.InstallDir, pathColor))
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
