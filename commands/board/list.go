package board

import (
	"fmt"
	"os"
	"time"

	"github.com/bcmi-labs/arduino-cli/arduino/cores/packagemanager"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/codeclysm/cc"

	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"

	"github.com/arduino/board-discovery"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   "List connected boards.",
		Long:    "Detects and displays a list of connected boards to the current computer.",
		Example: "arduino board list --timeout 10s",
		Args:    cobra.NoArgs,
		Run:     runListCommand,
	}
	listCommand.Flags().StringVar(&listFlags.timeout, "timeout", "5s", "The timeout of the search of connected devices, try to high it if your board is not found (e.g. to 10s).")
	return listCommand
}

var listFlags struct {
	timeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.
}

// runListCommand detects and lists the connected arduino boards
// (either via serial or network ports).
func runListCommand(cmd *cobra.Command, args []string) {
	pm := commands.InitPackageManager()

	if err := pm.LoadHardware(); err != nil {
		formatter.PrintError(err, "Error loading hardware files.")
		os.Exit(commands.ErrCoreConfig)
	}

	monitor := discovery.New(time.Millisecond)
	monitor.Start()
	duration, err := time.ParseDuration(listFlags.timeout)
	if err != nil {
		duration = time.Second * 5
	}
	if formatter.IsCurrentFormat("text") {
		stoppable := cc.Run(func(stop chan struct{}) {
			for {
				select {
				case <-stop:
					fmt.Print("\r              \r")
					return
				default:
					fmt.Print("\rDiscovering.  ")
					time.Sleep(time.Millisecond * 500)
					fmt.Print("\rDiscovering.. ")
					time.Sleep(time.Millisecond * 500)
					fmt.Print("\rDiscovering...")
					time.Sleep(time.Millisecond * 500)
				}
			}
		})

		fmt.Print("\r")

		time.Sleep(duration)
		stoppable.Stop()
		<-stoppable.Stopped
	} else {
		time.Sleep(duration)
	}

	formatter.Print(NewBoardList(pm, monitor))

	//monitor.Stop() //If called will slow like 1sec the program to close after print, with the same result (tested).
	// it closes ungracefully, but at the end of the command we can't have races.
}

// NewBoardList returns a new board list by adding discovered boards from the board list and a monitor.
func NewBoardList(pm *packagemanager.PackageManager, monitor *discovery.Monitor) *output.BoardList {
	if monitor == nil {
		return nil
	}

	serialDevices := monitor.Serial()
	networkDevices := monitor.Network()
	ret := &output.BoardList{
		SerialBoards:  make([]output.SerialBoardListItem, 0, len(serialDevices)),
		NetworkBoards: make([]output.NetworkBoardListItem, 0, len(networkDevices)),
	}

	for _, item := range serialDevices {
		boards := pm.FindBoardsWithVidPid(item.VendorID, item.ProductID)
		if len(boards) == 0 {
			// skip it if not recognized
			continue
		}

		board := boards[0]
		ret.SerialBoards = append(ret.SerialBoards, output.SerialBoardListItem{
			Name:  board.Name(),
			Fqbn:  board.FQBN(),
			Port:  item.Port,
			UsbID: fmt.Sprintf("%s:%s - %s", item.VendorID[2:], item.ProductID[2:], item.SerialNumber),
		})
	}

	for _, item := range networkDevices {
		boards := pm.FindBoardsWithID(item.Name)
		if len(boards) == 0 {
			// skip it if not recognized
			continue
		}

		board := boards[0]
		ret.NetworkBoards = append(ret.NetworkBoards, output.NetworkBoardListItem{
			Name:     board.Name(),
			Fqbn:     board.FQBN(),
			Location: fmt.Sprintf("%s:%d", item.Address, item.Port),
		})
	}
	return ret
}
