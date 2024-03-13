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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	srv_commands "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/version"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	tr           = i18n.Tr
	daemonize    bool
	debug        bool
	debugFile    string
	debugFilters []string
)

// NewCommand created a new `daemon` command
func NewCommand() *cobra.Command {
	daemonCommand := &cobra.Command{
		Use:     "daemon",
		Short:   tr("Run as a daemon on port: %s", configuration.Settings.GetString("daemon.port")),
		Long:    tr("Running as a daemon the initialization of cores and libraries is done only once."),
		Example: "  " + os.Args[0] + " daemon",
		Args:    cobra.NoArgs,
		Run:     runDaemonCommand,
	}
	daemonCommand.PersistentFlags().String("port", "", tr("The TCP port the daemon will listen to"))
	configuration.Settings.BindPFlag("daemon.port", daemonCommand.PersistentFlags().Lookup("port"))
	daemonCommand.Flags().BoolVar(&daemonize, "daemonize", false, tr("Do not terminate daemon process if the parent process dies"))
	daemonCommand.Flags().BoolVar(&debug, "debug", false, tr("Enable debug logging of gRPC calls"))
	daemonCommand.Flags().StringVar(&debugFile, "debug-file", "", tr("Append debug logging to the specified file"))
	daemonCommand.Flags().StringSliceVar(&debugFilters, "debug-filter", []string{}, tr("Display only the provided gRPC calls"))
	return daemonCommand
}

func runDaemonCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino-cli daemon`")

	// Bundled libraries support is enabled by default when running as a daemon
	configuration.Settings.SetDefault("directories.builtin.Libraries", configuration.GetDefaultBuiltinLibrariesDir())

	port := configuration.Settings.GetString("daemon.port")
	gRPCOptions := []grpc.ServerOption{}
	if debugFile != "" {
		if !debug {
			feedback.Fatal(tr("The flag --debug-file must be used with --debug."), feedback.ErrBadArgument)
		}
	}
	if debug {
		if debugFile != "" {
			outFile := paths.New(debugFile)
			f, err := outFile.Append()
			if err != nil {
				feedback.Fatal(tr("Error opening debug logging file: %s", err), feedback.ErrGeneric)
			}
			defer f.Close()
			debugStdOut = f
		} else {
			if out, _, err := feedback.DirectStreams(); err != nil {
				feedback.Fatal(tr("Can't write debug log: %s", err), feedback.ErrBadArgument)
			} else {
				debugStdOut = out
			}
		}
		gRPCOptions = append(gRPCOptions,
			grpc.UnaryInterceptor(unaryLoggerInterceptor),
			grpc.StreamInterceptor(streamLoggerInterceptor),
		)
	}
	s := grpc.NewServer(gRPCOptions...)
	// Set specific user-agent for the daemon
	configuration.Settings.Set("network.user_agent_ext", "daemon")

	// register the commands service
	srv_commands.RegisterArduinoCoreServiceServer(s, &commands.ArduinoCoreServerImpl{
		VersionString: version.VersionInfo.VersionString,
	})

	if !daemonize {
		// When parent process ends terminate also the daemon
		go feedback.ExitWhenParentProcessEnds()
	}

	ip := "127.0.0.1"
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		// Invalid port, such as "Foo"
		var dnsError *net.DNSError
		if errors.As(err, &dnsError) {
			feedback.Fatal(tr("Failed to listen on TCP port: %[1]s. %[2]s is unknown name.", port, dnsError.Name), feedback.ErrBadTCPPortArgument)
		}
		// Invalid port number, such as -1
		var addrError *net.AddrError
		if errors.As(err, &addrError) {
			feedback.Fatal(tr("Failed to listen on TCP port: %[1]s. %[2]s is an invalid port.", port, addrError.Addr), feedback.ErrBadTCPPortArgument)
		}
		// Port is already in use
		var syscallErr *os.SyscallError
		if errors.As(err, &syscallErr) && errors.Is(syscallErr.Err, syscall.EADDRINUSE) {
			feedback.Fatal(tr("Failed to listen on TCP port: %s. Address already in use.", port), feedback.ErrFailedToListenToTCPPort)
		}
		feedback.Fatal(tr("Failed to listen on TCP port: %[1]s. Unexpected error: %[2]v", port, err), feedback.ErrFailedToListenToTCPPort)
	}

	// We need to retrieve the port used only if the user did not specify it
	// and let the OS choose it randomly, in all other cases we already know
	// which port is used.
	if port == "0" {
		address := lis.Addr()
		split := strings.Split(address.String(), ":")

		if len(split) <= 1 {
			feedback.Fatal(tr("Invalid TCP address: port is missing"), feedback.ErrBadTCPPortArgument)
		}

		port = split[1]
	}

	feedback.PrintResult(daemonResult{
		IP:   ip,
		Port: port,
	})

	if err := s.Serve(lis); err != nil {
		feedback.Fatal(fmt.Sprintf("Failed to serve: %v", err), feedback.ErrFailedToListenToTCPPort)
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
	j, _ := json.Marshal(r)
	return fmt.Sprintln(tr("Daemon is now listening on %s:%s", r.IP, r.Port)) + fmt.Sprintln(string(j))
}
