// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package integrationtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/settings/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/fatih/color"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// FindRepositoryRootPath returns the repository root path
func FindRepositoryRootPath(t *testing.T) *paths.Path {
	repoRootPath, err := paths.Getwd()
	require.NoError(t, err)
	for !repoRootPath.Join(".git").Exist() {
		require.Contains(t, repoRootPath.String(), "arduino-cli", "Error searching for repository root path")
		repoRootPath = repoRootPath.Parent()
	}
	return repoRootPath
}

// FindArduinoCLIPath returns the path to the arduino-cli executable
func FindArduinoCLIPath(t *testing.T) *paths.Path {
	return FindRepositoryRootPath(t).Join("arduino-cli")
}

// CreateArduinoCLIWithEnvironment performs the minimum amount of actions
// to build the default test environment.
func CreateArduinoCLIWithEnvironment(t *testing.T) (*Environment, *ArduinoCLI) {
	env := NewEnvironment(t)

	cli := NewArduinoCliWithinEnvironment(env, &ArduinoCLIConfig{
		ArduinoCLIPath:         FindArduinoCLIPath(t),
		UseSharedStagingFolder: true,
	})

	return env, cli
}

// ArduinoCLI is an Arduino CLI client.
type ArduinoCLI struct {
	path                 *paths.Path
	t                    *require.Assertions
	proc                 *executils.Process
	cliEnvVars           map[string]string
	cliConfigPath        *paths.Path
	stagingDir           *paths.Path
	dataDir              *paths.Path
	sketchbookDir        *paths.Path
	workingDir           *paths.Path
	daemonAddr           string
	daemonConn           *grpc.ClientConn
	daemonClient         commands.ArduinoCoreServiceClient
	daemonSettingsClient settings.SettingsServiceClient
}

// ArduinoCLIConfig is the configuration of the ArduinoCLI client
type ArduinoCLIConfig struct {
	ArduinoCLIPath         *paths.Path
	UseSharedStagingFolder bool
}

// NewArduinoCliWithinEnvironment creates a new Arduino CLI client inside the given environment.
func NewArduinoCliWithinEnvironment(env *Environment, config *ArduinoCLIConfig) *ArduinoCLI {
	color.NoColor = false
	cli := &ArduinoCLI{
		path:          config.ArduinoCLIPath,
		t:             require.New(env.T()),
		dataDir:       env.RootDir().Join("A"),
		sketchbookDir: env.RootDir().Join("Arduino"),
		stagingDir:    env.RootDir().Join("Arduino15/staging"),
		workingDir:    env.RootDir(),
	}
	if config.UseSharedStagingFolder {
		sharedDir := env.SharedDownloadsDir()
		cli.stagingDir = sharedDir.Lock()
		env.RegisterCleanUpCallback(func() {
			sharedDir.Unlock()
		})
	}

	cli.cliEnvVars = map[string]string{
		"LANG":                   "en",
		"ARDUINO_DATA_DIR":       cli.dataDir.String(),
		"ARDUINO_DOWNLOADS_DIR":  cli.stagingDir.String(),
		"ARDUINO_SKETCHBOOK_DIR": cli.sketchbookDir.String(),
		"ARDUINO_BUILD_CACHE_COMPILATIONS_BEFORE_PURGE": "0",
	}
	env.RegisterCleanUpCallback(cli.CleanUp)
	return cli
}

// CleanUp closes the Arduino CLI client.
func (cli *ArduinoCLI) CleanUp() {
	if cli.proc != nil {
		cli.daemonConn.Close()
		cli.proc.Kill()
		cli.proc.Wait()
	}
}

// DataDir returns the data directory
func (cli *ArduinoCLI) DataDir() *paths.Path {
	return cli.dataDir
}

// SketchbookDir returns the sketchbook directory
func (cli *ArduinoCLI) SketchbookDir() *paths.Path {
	return cli.sketchbookDir
}

// WorkingDir returns the working directory
func (cli *ArduinoCLI) WorkingDir() *paths.Path {
	return cli.workingDir
}

// DownloadDir returns the download directory
func (cli *ArduinoCLI) DownloadDir() *paths.Path {
	return cli.stagingDir
}

// SetWorkingDir sets a new working directory
func (cli *ArduinoCLI) SetWorkingDir(p *paths.Path) {
	cli.workingDir = p
}

// CopySketch copies a sketch inside the testing environment and returns its path
func (cli *ArduinoCLI) CopySketch(sketchName string) *paths.Path {
	p, err := paths.Getwd()
	cli.t.NoError(err)
	cli.t.NotNil(p)
	testSketch := p.Parent().Join("testdata", sketchName)
	sketchPath := cli.WorkingDir().Join(sketchName)
	err = testSketch.CopyDirTo(sketchPath)
	cli.t.NoError(err)
	return sketchPath
}

// Run executes the given arduino-cli command and returns the output.
func (cli *ArduinoCLI) Run(args ...string) ([]byte, []byte, error) {
	return cli.RunWithCustomEnv(cli.cliEnvVars, args...)
}

// GetDefaultEnv returns a copy of the default execution env used with the Run method.
func (cli *ArduinoCLI) GetDefaultEnv() map[string]string {
	res := map[string]string{}
	for k, v := range cli.cliEnvVars {
		res[k] = v
	}
	return res
}

// convertEnvForExecutils returns a string array made of "key=value" strings
// with (key,value) pairs obtained from the given map.
func (cli *ArduinoCLI) convertEnvForExecutils(env map[string]string) []string {
	envVars := []string{}
	for k, v := range env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}
	return envVars
}

// RunWithCustomEnv executes the given arduino-cli command with the given custom env and returns the output.
func (cli *ArduinoCLI) RunWithCustomEnv(env map[string]string, args ...string) ([]byte, []byte, error) {
	if cli.cliConfigPath != nil {
		args = append([]string{"--config-file", cli.cliConfigPath.String()}, args...)
	}
	fmt.Println(color.HiBlackString(">>> Running: ") + color.HiYellowString("%s %s", cli.path, strings.Join(args, " ")))
	cliProc, err := executils.NewProcessFromPath(cli.convertEnvForExecutils(env), cli.path, args...)
	cli.t.NoError(err)
	stdout, err := cliProc.StdoutPipe()
	cli.t.NoError(err)
	stderr, err := cliProc.StderrPipe()
	cli.t.NoError(err)
	_, err = cliProc.StdinPipe()
	cli.t.NoError(err)
	cliProc.SetDir(cli.WorkingDir().String())

	cli.t.NoError(cliProc.Start())

	var stdoutBuf, stderrBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if _, err := io.Copy(&stdoutBuf, io.TeeReader(stdout, os.Stdout)); err != nil {
			fmt.Println(color.HiBlackString("<<< stdout copy error:"), err)
		}
	}()
	go func() {
		defer wg.Done()
		if _, err := io.Copy(&stderrBuf, io.TeeReader(stderr, os.Stderr)); err != nil {
			fmt.Println(color.HiBlackString("<<< stderr copy error:"), err)
		}
	}()
	wg.Wait()
	cliErr := cliProc.Wait()
	fmt.Println(color.HiBlackString("<<< Run completed (err = %v)", cliErr))

	errBuf := stderrBuf.Bytes()
	cli.t.NotContains(string(errBuf), "panic: runtime error:", "arduino-cli panicked")

	return stdoutBuf.Bytes(), errBuf, cliErr
}

// StartDaemon starts the Arduino CLI daemon. It returns the address of the daemon.
func (cli *ArduinoCLI) StartDaemon(verbose bool) string {
	args := []string{"daemon", "--format", "json"}
	if cli.cliConfigPath != nil {
		args = append([]string{"--config-file", cli.cliConfigPath.String()}, args...)
	}
	if verbose {
		args = append(args, "-v", "--log-level", "debug")
	}
	cliProc, err := executils.NewProcessFromPath(cli.convertEnvForExecutils(cli.cliEnvVars), cli.path, args...)
	cli.t.NoError(err)
	stdout, err := cliProc.StdoutPipe()
	cli.t.NoError(err)
	stderr, err := cliProc.StderrPipe()
	cli.t.NoError(err)
	_, err = cliProc.StdinPipe()
	cli.t.NoError(err)

	cli.t.NoError(cliProc.Start())
	cli.proc = cliProc
	cli.daemonAddr = "127.0.0.1:50051"

	copy := func(dst io.Writer, src io.Reader) {
		buff := make([]byte, 1024)
		for {
			n, err := src.Read(buff)
			if err != nil {
				return
			}
			dst.Write([]byte(color.YellowString(string(buff[:n]))))
		}
	}
	go copy(os.Stdout, stdout)
	go copy(os.Stderr, stderr)
	conn, err := grpc.Dial(cli.daemonAddr, grpc.WithInsecure(), grpc.WithBlock())
	cli.t.NoError(err)
	cli.daemonConn = conn
	cli.daemonClient = commands.NewArduinoCoreServiceClient(conn)
	cli.daemonSettingsClient = settings.NewSettingsServiceClient(conn)
	return cli.daemonAddr
}

// ArduinoCLIInstance is an Arduino CLI gRPC instance.
type ArduinoCLIInstance struct {
	cli      *ArduinoCLI
	instance *commands.Instance
}

var logCallfMutex sync.Mutex

func logCallf(format string, a ...interface{}) {
	logCallfMutex.Lock()
	fmt.Print(color.HiRedString(format, a...))
	logCallfMutex.Unlock()
}

// Create calls the "Create" gRPC method.
func (cli *ArduinoCLI) Create() *ArduinoCLIInstance {
	logCallf(">>> Create()")
	resp, err := cli.daemonClient.Create(context.Background(), &commands.CreateRequest{})
	cli.t.NoError(err)
	logCallf(" -> %v\n", resp)
	return &ArduinoCLIInstance{
		cli:      cli,
		instance: resp.Instance,
	}
}

// SetValue calls the "SetValue" gRPC method.
func (cli *ArduinoCLI) SetValue(key, jsonData string) error {
	req := &settings.SetValueRequest{
		Key:      key,
		JsonData: jsonData,
	}
	logCallf(">>> SetValue(%+v)\n", req)
	_, err := cli.daemonSettingsClient.SetValue(context.Background(), req)
	return err
}

// Init calls the "Init" gRPC method.
func (inst *ArduinoCLIInstance) Init(profile string, sketchPath string, respCB func(*commands.InitResponse)) error {
	initReq := &commands.InitRequest{
		Instance:   inst.instance,
		Profile:    profile,
		SketchPath: sketchPath,
	}
	logCallf(">>> Init(%v)\n", initReq)
	initClient, err := inst.cli.daemonClient.Init(context.Background(), initReq)
	if err != nil {
		return err
	}
	for {
		msg, err := initClient.Recv()
		if err == io.EOF {
			logCallf("<<< Init EOF\n")
			return nil
		}
		if err != nil {
			return err
		}
		if respCB != nil {
			respCB(msg)
		}
	}
}

// BoardList calls the "BoardList" gRPC method.
func (inst *ArduinoCLIInstance) BoardList(timeout time.Duration) (*commands.BoardListResponse, error) {
	boardListReq := &commands.BoardListRequest{
		Instance: inst.instance,
		Timeout:  timeout.Milliseconds(),
	}
	logCallf(">>> BoardList(%v) -> ", boardListReq)
	resp, err := inst.cli.daemonClient.BoardList(context.Background(), boardListReq)
	logCallf("err=%v\n", err)
	return resp, err
}

// BoardListWatch calls the "BoardListWatch" gRPC method.
func (inst *ArduinoCLIInstance) BoardListWatch() (commands.ArduinoCoreService_BoardListWatchClient, error) {
	boardListWatchReq := &commands.BoardListWatchRequest{
		Instance: inst.instance,
	}
	logCallf(">>> BoardListWatch(%v)\n", boardListWatchReq)
	watcher, err := inst.cli.daemonClient.BoardListWatch(context.Background())
	if err != nil {
		return watcher, err
	}
	return watcher, watcher.Send(boardListWatchReq)
}

// PlatformInstall calls the "PlatformInstall" gRPC method.
func (inst *ArduinoCLIInstance) PlatformInstall(ctx context.Context, packager, arch, version string, skipPostInst bool) (commands.ArduinoCoreService_PlatformInstallClient, error) {
	installCl, err := inst.cli.daemonClient.PlatformInstall(ctx, &commands.PlatformInstallRequest{
		Instance:        inst.instance,
		PlatformPackage: packager,
		Architecture:    arch,
		Version:         version,
		SkipPostInstall: skipPostInst,
	})
	logCallf(">>> PlatformInstall(%v:%v %v)\n", packager, arch, version)
	return installCl, err
}

// Compile calls the "Compile" gRPC method.
func (inst *ArduinoCLIInstance) Compile(ctx context.Context, fqbn, sketchPath string) (commands.ArduinoCoreService_CompileClient, error) {
	compileCl, err := inst.cli.daemonClient.Compile(ctx, &commands.CompileRequest{
		Instance:   inst.instance,
		Fqbn:       fqbn,
		SketchPath: sketchPath,
		Verbose:    true,
	})
	logCallf(">>> Compile(%v %v)\n", fqbn, sketchPath)
	return compileCl, err
}

// LibraryList calls the "LibraryList" gRPC method.
func (inst *ArduinoCLIInstance) LibraryList(ctx context.Context, name, fqbn string, all, updatable bool) (*commands.LibraryListResponse, error) {
	req := &commands.LibraryListRequest{
		Instance:  inst.instance,
		Name:      name,
		Fqbn:      fqbn,
		All:       all,
		Updatable: updatable,
	}
	logCallf(">>> LibraryList(%v) -> ", req)
	resp, err := inst.cli.daemonClient.LibraryList(ctx, req)
	logCallf("err=%v\n", err)
	r, _ := json.MarshalIndent(resp, "    ", "  ")
	logCallf("<<< LibraryList resp: %s\n", string(r))
	return resp, err
}

// LibraryInstall calls the "LibraryInstall" gRPC method.
func (inst *ArduinoCLIInstance) LibraryInstall(ctx context.Context, name, version string, noDeps, noOverwrite, installAsBundled bool) (commands.ArduinoCoreService_LibraryInstallClient, error) {
	installLocation := commands.LibraryInstallLocation_LIBRARY_INSTALL_LOCATION_USER
	if installAsBundled {
		installLocation = commands.LibraryInstallLocation_LIBRARY_INSTALL_LOCATION_BUILTIN
	}
	req := &commands.LibraryInstallRequest{
		Instance:        inst.instance,
		Name:            name,
		Version:         version,
		NoDeps:          noDeps,
		NoOverwrite:     noOverwrite,
		InstallLocation: installLocation,
	}
	installCl, err := inst.cli.daemonClient.LibraryInstall(ctx, req)
	logCallf(">>> LibraryInstall(%+v)\n", req)
	return installCl, err
}

// LibraryUninstall calls the "LibraryUninstall" gRPC method.
func (inst *ArduinoCLIInstance) LibraryUninstall(ctx context.Context, name, version string) (commands.ArduinoCoreService_LibraryUninstallClient, error) {
	req := &commands.LibraryUninstallRequest{
		Instance: inst.instance,
		Name:     name,
		Version:  version,
	}
	installCl, err := inst.cli.daemonClient.LibraryUninstall(ctx, req)
	logCallf(">>> LibraryUninstall(%+v)\n", req)
	return installCl, err
}

// UpdateIndex calls the "UpdateIndex" gRPC method.
func (inst *ArduinoCLIInstance) UpdateIndex(ctx context.Context, ignoreCustomPackages bool) (commands.ArduinoCoreService_UpdateIndexClient, error) {
	req := &commands.UpdateIndexRequest{
		Instance:                   inst.instance,
		IgnoreCustomPackageIndexes: ignoreCustomPackages,
	}
	updCl, err := inst.cli.daemonClient.UpdateIndex(ctx, req)
	logCallf(">>> UpdateIndex(%+v)\n", req)
	return updCl, err
}
