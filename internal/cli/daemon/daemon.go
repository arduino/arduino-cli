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
	"path/filepath"
	"strings"
	"syscall"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
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
func NewCommand(srv rpc.ArduinoCoreServiceServer, settings *rpc.Configuration) *cobra.Command {
	var daemonPort string
	daemonCommand := &cobra.Command{
		Use:     "daemon",
		Short:   i18n.Tr("Run the Arduino CLI as a gRPC daemon."),
		Example: "  " + os.Args[0] + " daemon",
		Args:    cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			// Bundled libraries support is enabled by default when running as a daemon
			if settings.GetDirectories().GetBuiltin().GetLibraries() == "" {
				defaultBuiltinLibDir := filepath.Join(settings.GetDirectories().GetData(), "libraries")
				_, err := srv.SettingsSetValue(cmd.Context(), &rpc.SettingsSetValueRequest{
					Key:          "directories.builtin.libraries",
					ValueFormat:  "cli",
					EncodedValue: defaultBuiltinLibDir,
				})
				if err != nil {
					// Should never happen...
					panic("Failed to set default value for directories.builtin.libraries: " + err.Error())
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			runDaemonCommand(srv, daemonPort)
		},
	}
	defaultDaemonPort := settings.GetDaemon().GetPort()

	daemonCommand.Flags().StringVar(&daemonPort,
		"port", defaultDaemonPort,
		i18n.Tr("The TCP port the daemon will listen to"))
	daemonCommand.Flags().BoolVar(&daemonize,
		"daemonize", false,
		i18n.Tr("Do not terminate daemon process if the parent process dies"))
	daemonCommand.Flags().BoolVar(&debug,
		"debug", false,
		i18n.Tr("Enable debug logging of gRPC calls"))
	daemonCommand.Flags().StringVar(&debugFile,
		"debug-file", "",
		i18n.Tr("Append debug logging to the specified file"))
	daemonCommand.Flags().StringSliceVar(&debugFilters,
		"debug-filter", []string{},
		i18n.Tr("Display only the provided gRPC calls"))
	return daemonCommand
}

func runDaemonCommand(srv rpc.ArduinoCoreServiceServer, daemonPort string) {
	logrus.Info("Executing `arduino-cli daemon`")

	gRPCOptions := []grpc.ServerOption{}
	if debugFile != "" {
		if !debug {
			feedback.Fatal(i18n.Tr("The flag --debug-file must be used with --debug."), feedback.ErrBadArgument)
		}
	}
	if debug {
		if debugFile != "" {
			outFile := paths.New(debugFile)
			f, err := outFile.Append()
			if err != nil {
				feedback.Fatal(i18n.Tr("Error opening debug logging file: %s", err), feedback.ErrGeneric)
			}
			defer f.Close()
			debugStdOut = f
		} else {
			if out, _, err := feedback.DirectStreams(); err != nil {
				feedback.Fatal(i18n.Tr("Can't write debug log: %s", err), feedback.ErrBadArgument)
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

	// register the commands service
	rpc.RegisterArduinoCoreServiceServer(s, srv)

	if !daemonize {
		// When parent process ends terminate also the daemon
		go feedback.ExitWhenParentProcessEnds()
	}

	daemonIP := "127.0.0.1"
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", daemonIP, daemonPort))
	if err != nil {
		// Invalid port, such as "Foo"
		var dnsError *net.DNSError
		if errors.As(err, &dnsError) {
			feedback.Fatal(i18n.Tr("Failed to listen on TCP port: %[1]s. %[2]s is unknown name.", daemonPort, dnsError.Name), feedback.ErrBadTCPPortArgument)
		}
		// Invalid port number, such as -1
		var addrError *net.AddrError
		if errors.As(err, &addrError) {
			feedback.Fatal(i18n.Tr("Failed to listen on TCP port: %[1]s. %[2]s is an invalid port.", daemonPort, addrError.Addr), feedback.ErrBadTCPPortArgument)
		}
		// Port is already in use
		var syscallErr *os.SyscallError
		if errors.As(err, &syscallErr) && errors.Is(syscallErr.Err, syscall.EADDRINUSE) {
			feedback.Fatal(i18n.Tr("Failed to listen on TCP port: %s. Address already in use.", daemonPort), feedback.ErrFailedToListenToTCPPort)
		}
		feedback.Fatal(i18n.Tr("Failed to listen on TCP port: %[1]s. Unexpected error: %[2]v", daemonPort, err), feedback.ErrFailedToListenToTCPPort)
	}

	// We need to retrieve the port used only if the user did not specify it
	// and let the OS choose it randomly, in all other cases we already know
	// which port is used.
	if daemonPort == "0" {
		address := lis.Addr()
		split := strings.Split(address.String(), ":")

		if len(split) <= 1 {
			feedback.Fatal(i18n.Tr("Invalid TCP address: port is missing"), feedback.ErrBadTCPPortArgument)
		}

		daemonPort = split[1]
	}

	feedback.PrintResult(daemonResult{
		IP:   daemonIP,
		Port: daemonPort,
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
	return fmt.Sprintln(i18n.Tr("Daemon is now listening on %s:%s", r.IP, r.Port)) + fmt.Sprintln(string(j))
}
