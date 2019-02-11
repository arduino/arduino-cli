//

package daemon

//go:generate protoc -I arduino --go_out=plugins=grpc:arduino arduino/arduino.proto

import (
	"fmt"
	"log"
	"net"

	"github.com/arduino/arduino-cli/cli"
	pb "github.com/arduino/arduino-cli/daemon/arduino"
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
	pb.RegisterArduinoCoreServer(s, &daemon{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	fmt.Println("Done serving")
}

// daemon is used to implement the Arduino Core Service.
type daemon struct{}
