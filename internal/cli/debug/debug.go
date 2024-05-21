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

package debug

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/table"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewCommand created a new `upload` command
func NewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var (
		fqbnArg     arguments.Fqbn
		portArgs    arguments.Port
		profileArg  arguments.Profile
		interpreter string
		importDir   string
		printInfo   bool
		programmer  arguments.Programmer
	)

	debugCommand := &cobra.Command{
		Use:     "debug",
		Short:   i18n.Tr("Debug Arduino sketches."),
		Long:    i18n.Tr("Debug Arduino sketches. (this command opens an interactive gdb session)"),
		Example: "  " + os.Args[0] + " debug -b arduino:samd:mkr1000 -P atmel_ice /home/user/Arduino/MySketch",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runDebugCommand(cmd.Context(), srv, args, &portArgs, &fqbnArg, interpreter, importDir, &programmer, printInfo, &profileArg)
		},
	}

	debugCommand.AddCommand(newDebugCheckCommand(srv))
	fqbnArg.AddToCommand(debugCommand, srv)
	portArgs.AddToCommand(debugCommand, srv)
	programmer.AddToCommand(debugCommand, srv)
	profileArg.AddToCommand(debugCommand, srv)
	debugCommand.Flags().StringVar(&interpreter, "interpreter", "console", i18n.Tr("Debug interpreter e.g.: %s", "console, mi, mi1, mi2, mi3"))
	debugCommand.Flags().StringVarP(&importDir, "input-dir", "", "", i18n.Tr("Directory containing binaries for debug."))
	debugCommand.Flags().BoolVarP(&printInfo, "info", "I", false, i18n.Tr("Show metadata about the debug session instead of starting the debugger."))

	return debugCommand
}

func runDebugCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string, portArgs *arguments.Port, fqbnArg *arguments.Fqbn,
	interpreter string, importDir string, programmer *arguments.Programmer, printInfo bool, profileArg *arguments.Profile) {
	logrus.Info("Executing `arduino-cli debug`")

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

	fqbn, port := arguments.CalculateFQBNAndPort(ctx, portArgs, fqbnArg, inst, srv, sk.GetDefaultFqbn(), sk.GetDefaultPort(), sk.GetDefaultProtocol())

	prog := profile.GetProgrammer()
	if prog == "" || programmer.GetProgrammer() != "" {
		prog = programmer.String(ctx, inst, srv, fqbn)
	}
	if prog == "" {
		prog = sk.GetDefaultProgrammer()
	}

	debugConfigRequested := &rpc.GetDebugConfigRequest{
		Instance:    inst,
		Fqbn:        fqbn,
		SketchPath:  sketchPath.String(),
		Port:        port,
		Interpreter: interpreter,
		ImportDir:   importDir,
		Programmer:  prog,
	}

	if printInfo {

		if res, err := commands.GetDebugConfig(ctx, debugConfigRequested); err != nil {
			errcode := feedback.ErrBadArgument
			if errors.Is(err, &cmderrors.MissingProgrammerError{}) {
				errcode = feedback.ErrMissingProgrammer
			}
			feedback.Fatal(i18n.Tr("Error getting Debug info: %v", err), errcode)
		} else {
			feedback.PrintResult(newDebugInfoResult(res))
		}

	} else {

		// Intercept SIGINT and forward them to debug process
		ctrlc := make(chan os.Signal, 1)
		signal.Notify(ctrlc, os.Interrupt)

		in, out, err := feedback.InteractiveStreams()
		if err != nil {
			feedback.FatalError(err, feedback.ErrBadArgument)
		}
		if _, err := commands.Debug(ctx, debugConfigRequested, in, out, ctrlc); err != nil {
			errcode := feedback.ErrGeneric
			if errors.Is(err, &cmderrors.MissingProgrammerError{}) {
				errcode = feedback.ErrMissingProgrammer
			}
			feedback.Fatal(i18n.Tr("Error during Debug: %v", err), errcode)
		}

	}
}

type debugInfoResult struct {
	Executable      string         `json:"executable,omitempty"`
	Toolchain       string         `json:"toolchain,omitempty"`
	ToolchainPath   string         `json:"toolchain_path,omitempty"`
	ToolchainPrefix string         `json:"toolchain_prefix,omitempty"`
	ToolchainConfig any            `json:"toolchain_configuration,omitempty"`
	Server          string         `json:"server,omitempty"`
	ServerPath      string         `json:"server_path,omitempty"`
	ServerConfig    any            `json:"server_configuration,omitempty"`
	SvdFile         string         `json:"svd_file,omitempty"`
	CustomConfigs   map[string]any `json:"custom_configs,omitempty"`
	Programmer      string         `json:"programmer"`
}

type openOcdServerConfigResult struct {
	Path       string   `json:"path,omitempty"`
	ScriptsDir string   `json:"scripts_dir,omitempty"`
	Scripts    []string `json:"scripts,omitempty"`
}

func newDebugInfoResult(info *rpc.GetDebugConfigResponse) *debugInfoResult {
	var toolchainConfig interface{}
	var serverConfig interface{}
	switch info.GetServer() {
	case "openocd":
		var openocdConf rpc.DebugOpenOCDServerConfiguration
		if err := info.GetServerConfiguration().UnmarshalTo(&openocdConf); err != nil {
			feedback.Fatal(i18n.Tr("Error during Debug: %v", err), feedback.ErrGeneric)
		}
		serverConfig = &openOcdServerConfigResult{
			Path:       openocdConf.GetPath(),
			ScriptsDir: openocdConf.GetScriptsDir(),
			Scripts:    openocdConf.GetScripts(),
		}
	}
	customConfigs := map[string]any{}
	for id, configJson := range info.GetCustomConfigs() {
		var config any
		if err := json.Unmarshal([]byte(configJson), &config); err == nil {
			customConfigs[id] = config
		}
	}
	return &debugInfoResult{
		Executable:      info.GetExecutable(),
		Toolchain:       info.GetToolchain(),
		ToolchainPath:   info.GetToolchainPath(),
		ToolchainPrefix: info.GetToolchainPrefix(),
		ToolchainConfig: toolchainConfig,
		Server:          info.GetServer(),
		ServerPath:      info.GetServerPath(),
		ServerConfig:    serverConfig,
		SvdFile:         info.GetSvdFile(),
		CustomConfigs:   customConfigs,
		Programmer:      info.GetProgrammer(),
	}
}

func (r *debugInfoResult) Data() interface{} {
	return r
}

func (r *debugInfoResult) String() string {
	t := table.New()
	green := color.New(color.FgHiGreen)
	dimGreen := color.New(color.FgGreen)
	t.AddRow(i18n.Tr("Executable to debug"), table.NewCell(r.Executable, green))
	t.AddRow(i18n.Tr("Toolchain type"), table.NewCell(r.Toolchain, green))
	t.AddRow(i18n.Tr("Toolchain path"), table.NewCell(r.ToolchainPath, dimGreen))
	t.AddRow(i18n.Tr("Toolchain prefix"), table.NewCell(r.ToolchainPrefix, dimGreen))
	if r.SvdFile != "" {
		t.AddRow(i18n.Tr("SVD file path"), table.NewCell(r.SvdFile, dimGreen))
	}
	switch r.Toolchain {
	case "gcc":
		// no options available at the moment...
	default:
	}
	t.AddRow(i18n.Tr("Server type"), table.NewCell(r.Server, green))
	t.AddRow(i18n.Tr("Server path"), table.NewCell(r.ServerPath, dimGreen))

	switch r.Server {
	case "openocd":
		t.AddRow(i18n.Tr("Configuration options for %s", r.Server))
		openocdConf := r.ServerConfig.(*openOcdServerConfigResult)
		if openocdConf.Path != "" {
			t.AddRow(" - Path", table.NewCell(openocdConf.Path, dimGreen))
		}
		if openocdConf.ScriptsDir != "" {
			t.AddRow(" - Scripts Directory", table.NewCell(openocdConf.ScriptsDir, dimGreen))
		}
		for _, script := range openocdConf.Scripts {
			t.AddRow(" - Script", table.NewCell(script, dimGreen))
		}
	default:
	}
	if custom := r.CustomConfigs; custom != nil {
		for id, config := range custom {
			configJson, _ := json.MarshalIndent(config, "", "  ")
			t.AddRow(i18n.Tr("Custom configuration for %s:", id))
			return t.Render() + "  " + string(configJson)
		}
	}
	return t.Render()
}
