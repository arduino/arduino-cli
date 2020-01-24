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

package daemon

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"syscall"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/commands/daemon"
	srv_commands "github.com/arduino/arduino-cli/rpc/commands"
	srv_debug "github.com/arduino/arduino-cli/rpc/debug"
	srv_monitor "github.com/arduino/arduino-cli/rpc/monitor"
	srv_settings "github.com/arduino/arduino-cli/rpc/settings"
	"github.com/arduino/arduino-cli/telemetry"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// NewCommand created a new `daemon` command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "daemon",
		Short:   fmt.Sprintf("Run as a daemon on port %s", viper.GetString("daemon.port")),
		Long:    "Running as a daemon the initialization of cores and libraries is done only once.",
		Example: "  " + os.Args[0] + " daemon",
		Args:    cobra.NoArgs,
		Run:     runDaemonCommand,
	}
	cmd.PersistentFlags().String("port", "", "The TCP port the daemon will listen to")
	viper.BindPFlag("daemon.port", cmd.PersistentFlags().Lookup("port"))
	cmd.Flags().BoolVar(&daemonize, "daemonize", false, "Do not terminate daemon process if the parent process dies")
	return cmd
}

var daemonize bool

func runDaemonCommand(cmd *cobra.Command, args []string) {

	if viper.GetBool("telemetry.enabled") {
		telemetry.Activate("daemon", viper.GetString("installation.id"))
		defer telemetry.Engine.Flush()
	}

	port := viper.GetString("daemon.port")
	s := grpc.NewServer()

	// Compose user agent header
	headers := http.Header{
		"User-Agent": []string{
			fmt.Sprintf("%s/%s daemon (%s; %s; %s) Commit:%s",
				globals.VersionInfo.Application,
				globals.VersionInfo.VersionString,
				runtime.GOARCH,
				runtime.GOOS,
				runtime.Version(),
				globals.VersionInfo.Commit),
		},
	}
	// Register the commands service
	srv_commands.RegisterArduinoCoreServer(s, &daemon.ArduinoCoreServerImpl{
		DownloaderHeaders: headers,
		VersionString:     globals.VersionInfo.VersionString,
	})

	// Register the monitors service
	srv_monitor.RegisterMonitorServer(s, &daemon.MonitorService{})

	// Register the settings service
	srv_settings.RegisterSettingsServer(s, &daemon.SettingsService{})

	// Register the debug session service
	srv_debug.RegisterDebugServer(s, &daemon.DebugService{})

	if !daemonize {
		// When parent process ends terminate also the daemon
		go func() {
			// Stdin is closed when the controlling parent process ends
			_, _ = io.Copy(ioutil.Discard, os.Stdin)
			os.Exit(0)
		}()
	}

	logrus.Infof("Starting daemon on TCP port %s", port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		// Invalid port, such as "Foo"
		var dnsError *net.DNSError
		if errors.As(err, &dnsError) {
			feedback.Errorf("Failed to listen on TCP port: %s. %s is unknown name.", port, dnsError.Name)
			os.Exit(errorcodes.ErrCoreConfig)
		}
		// Invalid port number, such as -1
		var addrError *net.AddrError
		if errors.As(err, &addrError) {
			feedback.Errorf("Failed to listen on TCP port: %s. %s is an invalid port.", port, addrError.Addr)
			os.Exit(errorcodes.ErrCoreConfig)
		}
		// Port is already in use
		var syscallErr *os.SyscallError
		if errors.As(err, &syscallErr) && errors.Is(syscallErr.Err, syscall.EADDRINUSE) {
			feedback.Errorf("Failed to listen on TCP port: %s. Address already in use.", port)
			os.Exit(errorcodes.ErrNetwork)
		}
		feedback.Errorf("Failed to listen on TCP port: %s. Unexpected error: %v", port, err)
		os.Exit(errorcodes.ErrGeneric)
	}
	// This message will show up on the stdout of the daemon process so that gRPC clients know it is time to connect.
	logrus.Infof("Daemon is listening on TCP port %s...", port)
	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("Failed to serve: %v", err)
	}
}
