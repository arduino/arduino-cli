//

package daemon

import (
	"fmt"
	"log"
	"net"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/daemon"
	"github.com/arduino/arduino-cli/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// InitCommand initalize the command
func InitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "daemon",
		Short:   "Run as a daemon",
		Long:    "Running as a daemon the initialization of cores and libraries is done only once.",
		Example: "  " + cli.AppName + " daemon",
		Args:    cobra.NoArgs,
		Run:     runDaemonCommand,
		Hidden:  true,
	}
	return cmd
}

const (
	port = ":50051"
)

func runDaemonCommand(cmd *cobra.Command, args []string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	rpc.RegisterArduinoCoreServer(s, &daemon.ArduinoCoreServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	fmt.Println("Done serving")
}
