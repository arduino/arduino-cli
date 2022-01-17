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
	"os"
	"strings"
	"syscall"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/commands/daemon"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/logging"
	"github.com/arduino/arduino-cli/output"
	srv_commands "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	srv_debug "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/debug/v1"
	srv_monitor "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/monitor/v1"
	srv_settings "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/settings/v1"
	"github.com/segmentio/stats/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	tr           = i18n.Tr
	debugFilters []string

	daemonConfigFile string
)

// NewCommand created a new `daemon` command
func NewCommand() *cobra.Command {
	daemonCommand := &cobra.Command{
		Use:     "daemon",
		Short:   tr("Run as a daemon on specified IP and port"),
		Long:    tr("Running as a daemon multiple different client can use the same Arduino CLI process with different settings."),
		Example: "  " + os.Args[0] + " daemon",
		Args:    cobra.NoArgs,
		Run:     runDaemonCommand,
	}
	daemonCommand.Flags().String("ip", "127.0.0.1", tr("The IP the daemon will listen to"))
	daemonCommand.Flags().String("port", "50051", tr("The TCP port the daemon will listen to"))
	daemonCommand.Flags().Bool("daemonize", false, tr("Run daemon process in background"))
	daemonCommand.Flags().Bool("debug", false, tr("Enable debug logging of gRPC calls"))
	daemonCommand.Flags().StringSlice("debug-filter", []string{}, tr("Display only the provided gRPC calls when debug is enabled"))
	daemonCommand.Flags().Bool("metrics-enabled", false, tr("Enable local metrics collection"))
	daemonCommand.Flags().String("metrics-address", ":9090", tr("Metrics local address"))
	// Metrics for the time being are ignored and unused, might as well hide this setting
	// from the user since they would do nothing.
	daemonCommand.Flags().MarkHidden("metrics-enabled")
	daemonCommand.Flags().MarkHidden("metrics-address")

	daemonCommand.Flags().StringVar(&daemonConfigFile, "config-file", "", tr("The daemon config file (if not specified default values will be used)."))
	return daemonCommand
}

func runDaemonCommand(cmd *cobra.Command, args []string) {
	s, err := load(cmd, daemonConfigFile)
	if err != nil {
		feedback.Errorf(tr("Error reading daemon config file: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}

	noColor := s.NoColor || os.Getenv("NO_COLOR") != ""
	output.Setup(s.OutputFormat, noColor)

	if daemonConfigFile != "" {
		// Tell the user which config file we're using only after output setup
		feedback.Printf(tr("Using daemon config file %s", daemonConfigFile))
	}

	logging.Setup(
		s.Verbose,
		noColor,
		s.LogLevel,
		s.LogFile,
		s.LogFormat,
	)

	logrus.Info("Executing `arduino-cli daemon`")

	gRPCOptions := []grpc.ServerOption{}
	if s.Debug {
		debugFilters = s.DebugFilter
		gRPCOptions = append(gRPCOptions,
			grpc.UnaryInterceptor(unaryLoggerInterceptor),
			grpc.StreamInterceptor(streamLoggerInterceptor),
		)
	}
	server := grpc.NewServer(gRPCOptions...)
	// Set specific user-agent for the daemon
	configuration.Settings.Set("network.user_agent_ext", "daemon")

	// register the commands service
	srv_commands.RegisterArduinoCoreServiceServer(server, &daemon.ArduinoCoreServerImpl{
		VersionString: globals.VersionInfo.VersionString,
	})

	// Register the monitors service
	srv_monitor.RegisterMonitorServiceServer(server, &daemon.MonitorService{})

	// Register the settings service
	srv_settings.RegisterSettingsServiceServer(server, &daemon.SettingsService{})

	// Register the debug session service
	srv_debug.RegisterDebugServiceServer(server, &daemon.DebugService{})

	if !s.Daemonize {
		// When parent process ends terminate also the daemon
		go func() {
			// Stdin is closed when the controlling parent process ends
			_, _ = io.Copy(ioutil.Discard, os.Stdin)
			// Flush metrics stats (this is a no-op if metrics is disabled)
			stats.Flush()
			os.Exit(0)
		}()
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.IP, s.Port))
	if err != nil {
		// Invalid port, such as "Foo"
		var dnsError *net.DNSError
		if errors.As(err, &dnsError) {
			feedback.Errorf(tr("Failed to listen on TCP port: %[1]s. %[2]s is unknown name."), s.Port, dnsError.Name)
			os.Exit(errorcodes.ErrCoreConfig)
		}
		// Invalid port number, such as -1
		var addrError *net.AddrError
		if errors.As(err, &addrError) {
			feedback.Errorf(tr("Failed to listen on TCP port: %[1]s. %[2]s is an invalid port."), s.Port, addrError.Addr)
			os.Exit(errorcodes.ErrCoreConfig)
		}
		// Port is already in use
		var syscallErr *os.SyscallError
		if errors.As(err, &syscallErr) && errors.Is(syscallErr.Err, syscall.EADDRINUSE) {
			feedback.Errorf(tr("Failed to listen on TCP port: %s. Address already in use."), s.Port)
			os.Exit(errorcodes.ErrNetwork)
		}
		feedback.Errorf(tr("Failed to listen on TCP port: %[1]s. Unexpected error: %[2]v"), s.Port, err)
		os.Exit(errorcodes.ErrGeneric)
	}

	// We need to parse the port used only if the user let
	// us choose it randomly, in all other cases we already
	// know which is used.
	if s.Port == "0" {
		address := lis.Addr()
		split := strings.Split(address.String(), ":")

		if len(split) == 0 {
			feedback.Error(tr("Failed choosing port, address: %s", address))
		}

		s.Port = split[len(split)-1]
	}

	feedback.PrintResult(daemonResult{
		IP:   s.IP,
		Port: s.Port,
	})

	if err := server.Serve(lis); err != nil {
		logrus.Fatalf("Failed to serve: %v", err)
	}
}

type daemonResult struct {
	IP   string
	Port string
}

func (r daemonResult) Data() interface{} {
	return r
}

func (r daemonResult) String() string {
	return tr("Daemon is now listening on %s:%s", r.IP, r.Port)
}
