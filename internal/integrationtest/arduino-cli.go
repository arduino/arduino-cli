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
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/fatih/color"
	"github.com/stretchr/testify/require"
	"go.bug.st/testsuite"
	"google.golang.org/grpc"
)

// ArduinoCLI is an Arduino CLI client.
type ArduinoCLI struct {
	path          *paths.Path
	t             *require.Assertions
	proc          *executils.Process
	cliEnvVars    []string
	cliConfigPath *paths.Path
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
func NewArduinoCliWithinEnvironment(env *testsuite.Environment, config *ArduinoCLIConfig) *ArduinoCLI {
	color.NoColor = false
	cli := NewArduinoCli(env.T(), config)
	staging := env.SharedDownloadsDir()
	if !config.UseSharedStagingFolder {
		staging = env.RootDir().Join("arduino15/staging")
	}
	cli.cliEnvVars = []string{
		fmt.Sprintf("ARDUINO_DATA_DIR=%s", env.RootDir().Join("arduino15")),
		fmt.Sprintf("ARDUINO_DOWNLOADS_DIR=%s", staging),
		fmt.Sprintf("ARDUINO_SKETCHBOOK_DIR=%s", env.RootDir().Join("Arduino")),
	}
	env.RegisterCleanUpCallback(cli.CleanUp)
	return cli
}

// NewArduinoCli creates a new Arduino CLI client.
func NewArduinoCli(t *testing.T, config *ArduinoCLIConfig) *ArduinoCLI {
	return &ArduinoCLI{
		path: config.ArduinoCLIPath,
		t:    require.New(t),
	}
}

// CleanUp closes the Arduino CLI client.
func (cli *ArduinoCLI) CleanUp() {
	if cli.proc != nil {
		cli.proc.Kill()
		cli.proc.Wait()
	}
}

// Run executes the given arduino-cli command and returns the output.
func (cli *ArduinoCLI) Run(args ...string) ([]byte, []byte, error) {
	if cli.cliConfigPath != nil {
		args = append([]string{"--config-file", cli.cliConfigPath.String()}, args...)
	}
	fmt.Println(color.HiBlackString(">>> Running: ") + color.HiYellowString("%s %s", cli.path, strings.Join(args, " ")))
	cliProc, err := executils.NewProcessFromPath(cli.cliEnvVars, cli.path, args...)
	cli.t.NoError(err)
	stdout, err := cliProc.StdoutPipe()
	cli.t.NoError(err)
	stderr, err := cliProc.StderrPipe()
	cli.t.NoError(err)
	_, err = cliProc.StdinPipe()
	cli.t.NoError(err)

	cli.t.NoError(cliProc.Start())

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutCtx, stdoutCancel := context.WithCancel(context.Background())
	stderrCtx, stderrCancel := context.WithCancel(context.Background())
	go func() {
		io.Copy(&stdoutBuf, io.TeeReader(stdout, os.Stdout))
		stdoutCancel()
	}()
	go func() {
		io.Copy(&stderrBuf, io.TeeReader(stderr, os.Stderr))
		stderrCancel()
	}()
	cliErr := cliProc.Wait()
	<-stdoutCtx.Done()
	<-stderrCtx.Done()
	fmt.Println(color.HiBlackString("<<< Run completed (err = %v)", cliErr))

	return stdoutBuf.Bytes(), stderrBuf.Bytes(), cliErr
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
	cliProc, err := executils.NewProcessFromPath(cli.cliEnvVars, cli.path, args...)
	cli.t.NoError(err)
	stdout, err := cliProc.StdoutPipe()
	cli.t.NoError(err)
	stderr, err := cliProc.StderrPipe()
	cli.t.NoError(err)
	_, err = cliProc.StdinPipe()
	cli.t.NoError(err)

	cli.t.NoError(cliProc.Start())
	cli.proc = cliProc

	// var daemonAddr struct {
	// 	IP   string
	// 	Port string
	// }
	// dec := json.NewDecoder(stdout)
	// cli.t.NoError(dec.Decode(&daemonAddr))
	// cli.daemonAddr = daemonAddr.IP + ":" + daemonAddr.Port
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
