package board

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/codeclysm/cc"

	"github.com/bcmi-labs/arduino-modules/boards"

	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"

	"github.com/arduino/board-discovery"
	"github.com/spf13/cobra"
)

func init() {
	command.AddCommand(listCommand)
	listCommand.Flags().StringVar(&listFlags.timeout, "timeout", "5s", "The timeout of the search of connected devices, try to high it if your board is not found (e.g. to 10s).")
}

var listFlags struct {
	timeout string // Expressed in a parsable duration, is the timeout for the list and attach commands.
}

var listCommand = &cobra.Command{
	Use: "list",
	Run: runListCommand,
}

var validSerialBoardURIRegexp = regexp.MustCompile("(serial|tty)://.+")
var validNetworkBoardURIRegexp = regexp.MustCompile("(http(s)?|(tc|ud)p)://[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}:[0-9]{1,5}")

// runListCommand detects and lists the connected arduino boards
// (either via serial or network ports).
func runListCommand(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		formatter.PrintErrorMessage("Not accepting additional arguments.")
		os.Exit(commands.ErrBadCall)
	}

	packageFolder, err := common.GetDefaultPkgFolder()
	if err != nil {
		formatter.PrintError(err, "Cannot Parse Board Index file.")
		os.Exit(commands.ErrCoreConfig)
	}

	bs, err := boards.Find(packageFolder)
	if err != nil {
		formatter.PrintError(err, "Cannot Parse Board Index file.")
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

	formatter.Print(output.NewBoardList(bs, monitor))

	//monitor.Stop() //If called will slow like 1sec the program to close after print, with the same result (tested).
	// it closes ungracefully, but at the end of the command we can't have races.
}

// findSerialConnectedBoard find the board which is connected to the specified URI via serial port, using a monitor and a set of Boards
// for the matching.
func findSerialConnectedBoard(bs boards.Boards, monitor *discovery.Monitor, deviceURI *url.URL) *boards.Board {
	found := false
	location := deviceURI.Path
	var serialDevice discovery.SerialDevice
	for _, device := range monitor.Serial() {
		if device.Port == location {
			// Found the device !
			found = true
			serialDevice = *device
		}
	}
	if !found {
		formatter.PrintErrorMessage("No Supported board has been found at the specified board URI.")
		return nil
	}

	board := bs.ByVidPid(serialDevice.VendorID, serialDevice.ProductID)
	if board == nil {
		formatter.PrintErrorMessage("No Supported board has been found, try either install new cores or check your board URI.")
		os.Exit(commands.ErrGeneric)
	}

	formatter.Print("SUPPORTED BOARD FOUND:")
	formatter.Print(board.String())
	return board
}

// findNetworkConnectedBoard find the board which is connected to the specified URI on the network, using a monitor and a set of Boards
// for the matching.
func findNetworkConnectedBoard(bs boards.Boards, monitor *discovery.Monitor, deviceURI *url.URL) *boards.Board {
	found := false

	var networkDevice discovery.NetworkDevice

	for _, device := range monitor.Network() {
		if device.Address == deviceURI.Host &&
			fmt.Sprint(device.Port) == deviceURI.Port() {
			// Found the device !
			found = true
			networkDevice = *device
		}
	}
	if !found {
		formatter.PrintErrorMessage("No Supported board has been found at the specified board URI, try either install new cores or check your board URI.")
		os.Exit(commands.ErrGeneric)
	}

	formatter.Print("SUPPORTED BOARD FOUND:")
	return bs.ByID(networkDevice.Name)
}
