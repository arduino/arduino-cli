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
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"

	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/commands/daemon"
	srv_commands "github.com/arduino/arduino-cli/rpc/commands"
	srv_debug "github.com/arduino/arduino-cli/rpc/debug"
	srv_monitor "github.com/arduino/arduino-cli/rpc/monitor"
	srv_settings "github.com/arduino/arduino-cli/rpc/settings"
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
	port := viper.GetString("daemon.port")
	s := grpc.NewServer()

	// register the commands service
	headers := http.Header{"User-Agent": []string{
		fmt.Sprintf("%s/%s daemon (%s; %s; %s) Commit:%s",
			globals.VersionInfo.Application,
			globals.VersionInfo.VersionString,
			runtime.GOARCH, runtime.GOOS,
			runtime.Version(), globals.VersionInfo.Commit)}}
	srv_commands.RegisterArduinoCoreServer(s, &daemon.ArduinoCoreServerImpl{
		DownloaderHeaders: headers,
		VersionString:     globals.VersionInfo.VersionString,
	})

	// register the monitors service
	srv_monitor.RegisterMonitorServer(s, &daemon.MonitorService{})

	// register the settings service
	srv_settings.RegisterSettingsServer(s, &daemon.SettingsService{})

	// register the debug session service
	srv_debug.RegisterDebugServer(s, &daemon.DebugService{})

	if !daemonize {
		// When parent process ends terminate also the daemon
		go func() {
			// stdin is closed when the controlling parent process ends
			_, _ = io.Copy(ioutil.Discard, os.Stdin)
			os.Exit(0)
		}()
	}

	logrus.Infof("Starting daemon on TCP port %s", port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}
	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("failed to serve: %v", err)
	}
}
