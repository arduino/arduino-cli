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
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/fatih/color"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	path          *paths.Path
	t             *require.Assertions
	proc          *paths.Process
	stdIn         io.WriteCloser
	cliEnvVars    map[string]string
	cliConfigPath *paths.Path
	stagingDir    *paths.Path
	dataDir       *paths.Path
	sketchbookDir *paths.Path
	workingDir    *paths.Path
	daemonAddr    string
	daemonConn    *grpc.ClientConn
	daemonClient  commands.ArduinoCoreServiceClient
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
		"LANG":                                          "en",
		"ARDUINO_DIRECTORIES_DATA":                      cli.dataDir.String(),
		"ARDUINO_DIRECTORIES_DOWNLOADS":                 cli.stagingDir.String(),
		"ARDUINO_DIRECTORIES_USER":                      cli.sketchbookDir.String(),
		"ARDUINO_BUILD_CACHE_COMPILATIONS_BEFORE_PURGE": "0",
	}
	env.RegisterCleanUpCallback(cli.CleanUp)
	return cli
}

// CreateEnvForDaemon performs the minimum required operations to start the arduino-cli daemon.
// It returns a testsuite.Environment and an ArduinoCLI client to perform the integration tests.
// The Environment must be disposed by calling the CleanUp method via defer.
func CreateEnvForDaemon(t *testing.T) (*Environment, *ArduinoCLI) {
	env := NewEnvironment(t)

	cli := NewArduinoCliWithinEnvironment(env, &ArduinoCLIConfig{
		ArduinoCLIPath:         FindRepositoryRootPath(t).Join("arduino-cli"),
		UseSharedStagingFolder: true,
	})

	_ = cli.StartDaemon(false)
	return env, cli
}

// CleanUp closes the Arduino CLI client.
func (cli *ArduinoCLI) CleanUp() {
	if cli.proc != nil {
		cli.daemonConn.Close()
		cli.stdIn.Close()
		proc := cli.proc
		go func() {
			time.Sleep(time.Second)
			proc.Kill()
		}()
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

// RunWithContext executes the given arduino-cli command with the given context and returns the output.
// If the context is canceled, the command is killed.
func (cli *ArduinoCLI) RunWithContext(ctx context.Context, args ...string) ([]byte, []byte, error) {
	return cli.RunWithCustomEnvContext(ctx, cli.cliEnvVars, args...)
}

// GetDefaultEnv returns a copy of the default execution env used with the Run method.
func (cli *ArduinoCLI) GetDefaultEnv() map[string]string {
	res := map[string]string{}
	maps.Copy(res, cli.cliEnvVars)
	return res
}

// convertEnvForExecutils returns a string array made of "key=value" strings
// with (key,value) pairs obtained from the given map.
func (cli *ArduinoCLI) convertEnvForExecutils(env map[string]string) []string {
	envVars := []string{}
	for k, v := range env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}

	// Proxy code-coverage related env vars
	if gocoverdir := os.Getenv("INTEGRATION_GOCOVERDIR"); gocoverdir != "" {
		envVars = append(envVars, "GOCOVERDIR="+gocoverdir)
	}
	return envVars
}

// InstallMockedSerialDiscovery will replace the already installed serial-discovery
// with a mocked one.
func (cli *ArduinoCLI) InstallMockedSerialDiscovery(t *testing.T) {
	fmt.Println(color.BlueString("<<< Install mocked serial-discovery"))

	// Build mocked serial-discovery
	mockDir := FindRepositoryRootPath(t).Join("internal", "mock_serial_discovery")
	gobuild, err := paths.NewProcess(nil, "go", "build")
	require.NoError(t, err)
	gobuild.SetDirFromPath(mockDir)
	require.NoError(t, gobuild.Run(), "Building mocked serial-discovery")
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	mockBin := mockDir.Join("mock_serial_discovery" + ext)
	require.True(t, mockBin.Exist())
	fmt.Println(color.HiBlackString("    Build of mocked serial-discovery succeeded."))

	// Install it replacing the current serial discovery
	dataDir := cli.DataDir()
	require.NotNil(t, dataDir, "data dir missing")
	serialDiscoveries, err := dataDir.Join("packages", "builtin", "tools", "serial-discovery").ReadDirRecursiveFiltered(
		nil, paths.AndFilter(
			paths.FilterNames("serial-discovery"+ext),
			paths.FilterOutDirectories(),
		),
	)
	require.NoError(t, err, "scanning data dir for serial-discoveries")
	require.NotEmpty(t, serialDiscoveries, "no serial-discoveries found in data dir")
	for _, serialDiscovery := range serialDiscoveries {
		require.NoError(t, mockBin.CopyTo(serialDiscovery), "installing mocked serial discovery to %s", serialDiscovery)
		fmt.Println(color.HiBlackString("    Discovery installed in " + serialDiscovery.String()))
	}
}

// InstallMockedSerialMonitor will replace the already installed serial-monitor
// with a mocked one.
func (cli *ArduinoCLI) InstallMockedSerialMonitor(t *testing.T) {
	fmt.Println(color.BlueString("<<< Install mocked serial-monitor"))

	// Build mocked serial-monitor
	mockDir := FindRepositoryRootPath(t).Join("internal", "mock_serial_monitor")
	gobuild, err := paths.NewProcess(nil, "go", "build")
	require.NoError(t, err)
	gobuild.SetDirFromPath(mockDir)
	require.NoError(t, gobuild.Run(), "Building mocked serial-monitor")
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	mockBin := mockDir.Join("mock_serial_monitor" + ext)
	require.True(t, mockBin.Exist())
	fmt.Println(color.HiBlackString("    Build of mocked serial-monitor succeeded."))

	// Install it replacing the current serial monitor
	dataDir := cli.DataDir()
	require.NotNil(t, dataDir, "data dir missing")
	serialMonitors, err := dataDir.Join("packages", "builtin", "tools", "serial-monitor").ReadDirRecursiveFiltered(
		nil, paths.AndFilter(
			paths.FilterNames("serial-monitor"+ext),
			paths.FilterOutDirectories(),
		),
	)
	require.NoError(t, err, "scanning data dir for serial-monitor")
	require.NotEmpty(t, serialMonitors, "no serial-monitor found in data dir")
	for _, serialMonitor := range serialMonitors {
		require.NoError(t, mockBin.CopyTo(serialMonitor), "installing mocked serial monitor to %s", serialMonitor)
		fmt.Println(color.HiBlackString("    Monitor installed in " + serialMonitor.String()))
	}
}

// InstallMockedAvrdude will replace the already installed avrdude with a mocked one.
func (cli *ArduinoCLI) InstallMockedAvrdude(t *testing.T) {
	fmt.Println(color.BlueString("<<< Install mocked avrdude"))

	// Build mocked serial-discovery
	mockDir := FindRepositoryRootPath(t).Join("internal", "mock_avrdude")
	gobuild, err := paths.NewProcess(nil, "go", "build")
	require.NoError(t, err)
	gobuild.SetDirFromPath(mockDir)
	require.NoError(t, gobuild.Run(), "Building mocked avrdude")
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	mockBin := mockDir.Join("mock_avrdude" + ext)
	require.True(t, mockBin.Exist())
	fmt.Println(color.HiBlackString("    Build of mocked avrdude succeeded."))

	// Install it replacing the current avrdudes
	dataDir := cli.DataDir()
	require.NotNil(t, dataDir, "data dir missing")

	avrdudes, err := dataDir.Join("packages", "arduino", "tools", "avrdude").ReadDirRecursiveFiltered(
		nil, paths.AndFilter(
			paths.FilterNames("avrdude"+ext),
			paths.FilterOutDirectories(),
		),
	)
	require.NoError(t, err, "scanning data dir for avrdude(s)")
	require.NotEmpty(t, avrdudes, "no avrdude(s) found in data dir")
	for _, avrdude := range avrdudes {
		require.NoError(t, mockBin.CopyTo(avrdude), "installing mocked avrdude to %s", avrdude)
		fmt.Println(color.HiBlackString("    Mocked avrdude installed in " + avrdude.String()))
	}
}

// RunWithCustomEnv executes the given arduino-cli command with the given custom env and returns the output.
func (cli *ArduinoCLI) RunWithCustomEnv(env map[string]string, args ...string) ([]byte, []byte, error) {
	return cli.RunWithCustomEnvContext(context.Background(), env, args...)
}

// RunWithCustomEnv executes the given arduino-cli command with the given custom env and returns the output.
func (cli *ArduinoCLI) RunWithCustomEnvContext(ctx context.Context, env map[string]string, args ...string) ([]byte, []byte, error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	err := cli.run(ctx, &stdoutBuf, &stderrBuf, nil, env, args...)

	errBuf := stderrBuf.Bytes()
	cli.t.NotContains(string(errBuf), "panic: runtime error:", "arduino-cli panicked")

	return stdoutBuf.Bytes(), errBuf, err
}

// RunWithCustomInput executes the given arduino-cli command pushing the given input stream and returns the output.
func (cli *ArduinoCLI) RunWithCustomInput(in io.Reader, args ...string) ([]byte, []byte, error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	err := cli.run(context.Background(), &stdoutBuf, &stderrBuf, in, cli.cliEnvVars, args...)

	errBuf := stderrBuf.Bytes()
	cli.t.NotContains(string(errBuf), "panic: runtime error:", "arduino-cli panicked")

	return stdoutBuf.Bytes(), errBuf, err
}

func (cli *ArduinoCLI) run(ctx context.Context, stdoutBuff, stderrBuff io.Writer, stdinBuff io.Reader, env map[string]string, args ...string) (_err error) {
	if cli.cliConfigPath != nil {
		args = append([]string{"--config-file", cli.cliConfigPath.String()}, args...)
	}

	// Accumulate all output to terminal and spit-out all at once at the end of the test
	// This allows to correctly group test output when running t.Parallel() tests.
	terminalOut := new(bytes.Buffer)
	terminalErr := new(bytes.Buffer)

	// Github-actions workflow tags to fold log lines
	if os.Getenv("GITHUB_ACTIONS") != "" {
		fmt.Fprintf(terminalOut, "::group::Running %s\n", strings.Join(args, " "))
		defer fmt.Fprintln(terminalOut, "::endgroup::")
	}

	fmt.Fprintln(terminalOut, color.HiBlackString(">>> Running: ")+color.HiYellowString("%s %s %s", cli.path, strings.Join(args, " "), env))
	defer func() {
		fmt.Print(terminalOut.String())
		fmt.Print(terminalErr.String())
		fmt.Println(color.HiBlackString("<<< Run completed (err = %v)", _err))
	}()

	cliProc, err := paths.NewProcessFromPath(cli.convertEnvForExecutils(env), cli.path, args...)
	cli.t.NoError(err)
	stdout, err := cliProc.StdoutPipe()
	cli.t.NoError(err)
	stderr, err := cliProc.StderrPipe()
	cli.t.NoError(err)
	stdin, err := cliProc.StdinPipe()
	cli.t.NoError(err)
	cliProc.SetDir(cli.WorkingDir().String())

	cli.t.NoError(cliProc.Start())

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if stdoutBuff == nil {
			stdoutBuff = io.Discard
		}
		if _, err := io.Copy(stdoutBuff, io.TeeReader(stdout, terminalOut)); err != nil {
			fmt.Fprintln(terminalOut, color.HiBlackString("<<< stdout copy error:"), err)
		}
	}()
	go func() {
		defer wg.Done()
		if stderrBuff == nil {
			stderrBuff = io.Discard
		}
		if _, err := io.Copy(stderrBuff, io.TeeReader(stderr, terminalErr)); err != nil {
			fmt.Fprintln(terminalErr, color.HiBlackString("<<< stderr copy error:"), err)
		}
	}()
	if stdinBuff != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := io.Copy(stdin, stdinBuff); err != nil {
				fmt.Fprintln(terminalErr, color.HiBlackString("<<< stdin copy error:"), err)
			}
		}()
	}
	wg.Wait()
	return cliProc.WaitWithinContext(ctx)
}

// StartDaemon starts the Arduino CLI daemon. It returns the address of the daemon.
func (cli *ArduinoCLI) StartDaemon(verbose bool) string {
	args := []string{"daemon", "--json"}
	if cli.cliConfigPath != nil {
		args = append([]string{"--config-file", cli.cliConfigPath.String()}, args...)
	}
	if verbose {
		args = append(args, "-v", "--log-level", "debug")
	}
	cliProc, err := paths.NewProcessFromPath(cli.convertEnvForExecutils(cli.cliEnvVars), cli.path, args...)
	cli.t.NoError(err)
	stdout, err := cliProc.StdoutPipe()
	cli.t.NoError(err)
	stderr, err := cliProc.StderrPipe()
	cli.t.NoError(err)
	stdIn, err := cliProc.StdinPipe()
	cli.t.NoError(err)

	cli.t.NoError(cliProc.Start())
	cli.stdIn = stdIn
	cli.proc = cliProc
	cli.daemonAddr = "127.0.0.1:50051"

	_copy := func(dst io.Writer, src io.Reader) {
		buff := make([]byte, 1024)
		for {
			n, err := src.Read(buff)
			if err != nil {
				return
			}
			dst.Write([]byte(color.YellowString(string(buff[:n]))))
		}
	}
	go _copy(os.Stdout, stdout)
	go _copy(os.Stderr, stderr)

	// Await the CLI daemon to be ready
	var connErr error
	for retries := 5; retries > 0; retries-- {
		time.Sleep(time.Second)

		conn, err := grpc.NewClient(
			cli.daemonAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUserAgent("cli-test/0.0.0"),
		)
		if err != nil {
			connErr = err
			continue
		}
		cli.daemonConn = conn
		cli.daemonClient = commands.NewArduinoCoreServiceClient(conn)
		break
	}
	cli.t.NoError(connErr)
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
		instance: resp.GetInstance(),
	}
}

// SetValue calls the "SetValue" gRPC method.
func (cli *ArduinoCLI) SetValue(key, jsonData string) error {
	req := &commands.SettingsSetValueRequest{
		Key:          key,
		EncodedValue: jsonData,
	}
	logCallf(">>> SetValue(%+v)\n", req)
	_, err := cli.daemonClient.SettingsSetValue(context.Background(), req)
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
		if errors.Is(err, io.EOF) {
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
func (inst *ArduinoCLIInstance) BoardListWatch(ctx context.Context) (commands.ArduinoCoreService_BoardListWatchClient, error) {
	boardListWatchReq := &commands.BoardListWatchRequest{
		Instance: inst.instance,
	}
	logCallf(">>> BoardListWatch(%v)\n", boardListWatchReq)
	watcher, err := inst.cli.daemonClient.BoardListWatch(ctx, boardListWatchReq)
	if err != nil {
		return watcher, err
	}
	return watcher, nil
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
func (inst *ArduinoCLIInstance) Compile(ctx context.Context, fqbn, sketchPath string, warnings string) (commands.ArduinoCoreService_CompileClient, error) {
	compileCl, err := inst.cli.daemonClient.Compile(ctx, &commands.CompileRequest{
		Instance:   inst.instance,
		Fqbn:       fqbn,
		SketchPath: sketchPath,
		Verbose:    true,
		Warnings:   warnings,
	})
	logCallf(">>> Compile(%v %v warnings=%v)\n", fqbn, sketchPath, warnings)
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

// PlatformUpgrade calls the "PlatformUpgrade" gRPC method.
func (inst *ArduinoCLIInstance) PlatformUpgrade(ctx context.Context, packager, arch string, skipPostInst bool) (commands.ArduinoCoreService_PlatformUpgradeClient, error) {
	installCl, err := inst.cli.daemonClient.PlatformUpgrade(ctx, &commands.PlatformUpgradeRequest{
		Instance:        inst.instance,
		PlatformPackage: packager,
		Architecture:    arch,
		SkipPostInstall: skipPostInst,
	})
	logCallf(">>> PlatformUpgrade(%v:%v)\n", packager, arch)
	return installCl, err
}

// PlatformSearch calls the "PlatformSearch" gRPC method.
func (inst *ArduinoCLIInstance) PlatformSearch(ctx context.Context, args string, all bool) (*commands.PlatformSearchResponse, error) {
	req := &commands.PlatformSearchRequest{
		Instance:   inst.instance,
		SearchArgs: args,
	}
	logCallf(">>> PlatformSearch(%+v)\n", req)
	resp, err := inst.cli.daemonClient.PlatformSearch(ctx, req)
	return resp, err
}

// Monitor calls the "Monitor" gRPC method and sends the OpenRequest message.
func (inst *ArduinoCLIInstance) Monitor(ctx context.Context, port *commands.Port) (commands.ArduinoCoreService_MonitorClient, error) {
	req := &commands.MonitorRequest{}
	logCallf(">>> Monitor(%+v)\n", req)
	monitorClient, err := inst.cli.daemonClient.Monitor(ctx)
	if err != nil {
		return nil, err
	}
	err = monitorClient.Send(&commands.MonitorRequest{
		Message: &commands.MonitorRequest_OpenRequest{
			OpenRequest: &commands.MonitorPortOpenRequest{
				Instance: inst.instance,
				Port:     port,
			},
		},
	})
	return monitorClient, err
}

// Upload calls the "Upload" gRPC method.
func (inst *ArduinoCLIInstance) Upload(ctx context.Context, fqbn, sketchPath, port, protocol string) (commands.ArduinoCoreService_UploadClient, error) {
	uploadCl, err := inst.cli.daemonClient.Upload(ctx, &commands.UploadRequest{
		Instance:   inst.instance,
		Fqbn:       fqbn,
		SketchPath: sketchPath,
		Verbose:    true,
		Port: &commands.Port{
			Address:  port,
			Protocol: protocol,
		},
	})
	logCallf(">>> Upload(%v %v port/protocol=%s/%s)\n", fqbn, sketchPath, port, protocol)
	return uploadCl, err
}

// BoardIdentify calls the "BoardIdentify" gRPC method.
func (inst *ArduinoCLIInstance) BoardIdentify(ctx context.Context, props map[string]string, useCloudAPI bool) (*commands.BoardIdentifyResponse, error) {
	req := &commands.BoardIdentifyRequest{
		Instance:                            inst.instance,
		Properties:                          props,
		UseCloudApiForUnknownBoardDetection: useCloudAPI,
	}
	logCallf(">>> BoardIdentify(%+v)\n", req)
	resp, err := inst.cli.daemonClient.BoardIdentify(ctx, req)
	return resp, err
}

// NewSketch calls the "NewSketch" gRPC method.
func (inst *ArduinoCLIInstance) NewSketch(ctx context.Context, sketchName, sketchDir string, overwrite bool) (*commands.NewSketchResponse, error) {
	req := &commands.NewSketchRequest{
		SketchName: sketchName,
		SketchDir:  sketchDir,
		Overwrite:  overwrite,
	}
	logCallf(">>> NewSketch(%+v)\n", req)
	return inst.cli.daemonClient.NewSketch(ctx, req)
}
