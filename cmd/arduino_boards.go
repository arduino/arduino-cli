package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/codeclysm/cc"

	"github.com/bcmi-labs/arduino-modules/boards"
	"github.com/bcmi-labs/arduino-modules/sketches"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"

	"github.com/arduino/board-discovery"
	"github.com/spf13/cobra"
)

var arduinoBoardCmd = &cobra.Command{
	Use:   "board",
	Short: `Arduino board commands`,
	Long:  `Arduino board commands`,
	Example: `arduino board list                     # Lists all connected boards
arduino board attach --board serial:///dev/tty/ACM0 \
                     --sketch mySketch # Attaches a sketch to a board`,
}

var arduinoBoardListCmd = &cobra.Command{
	Use: "list",
	Run: executeBoardListCommand,
}

var arduinoBoardAttachCmd = &cobra.Command{
	Use:   "attach --sketch=[SKETCH-NAME] --board=[BOARD]",
	Short: `Attaches a board to a sketch`,
	Long:  `Attaches a board to a sketch`,
	Example: `arduino board attach --board serial:///dev/tty/ACM0 \
	             --sketch mySketch # Attaches a sketch to a board`,
	RunE: executeBoardAttachCommand,
}

// executeBoardListCommand detects and lists the connected arduino boards
// (either via serial or network ports).
func executeBoardListCommand(cmd *cobra.Command, args []string) {
	monitor := discovery.New(time.Millisecond)
	monitor.Start()
	duration, err := time.ParseDuration(arduinoBoardListFlags.SearchTimeout)
	if err != nil {
		duration = time.Second * 5
	}
	stoppable := cc.Run(func(stop chan struct{}) {
		for {
			select {
			case <-stop:
				fmt.Print("\r              ")
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

	time.Sleep(duration)
	stoppable.Stop()
	<-stoppable.Stopped
	fmt.Print("\r")
	formatter.Print(*monitor)
	//monitor.Stop() //If called will slow like 1sec the program to close after print, with the same result (tested).
	// it closes ungracefully, but at the end of the command we can't have races.
}

func executeBoardAttachCommand(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return errors.New("Not accepting additional arguments")
	}

	if arduinoBoardAttachFlags.SketchName == "" {
		return errors.New("No sketch name provided")
	}

	if arduinoBoardAttachFlags.BoardURI == "" {
		return errors.New("No board URI provided")
	}

	duration, err := time.ParseDuration(arduinoBoardListFlags.SearchTimeout)
	if err != nil {
		duration = time.Second * 5
	}

	monitor := discovery.New(time.Second)
	monitor.Start()

	time.Sleep(duration)

	homeFolder, err := common.GetDefaultArduinoHomeFolder()
	if err != nil {
		formatter.PrintErrorMessage("Cannot Parse Board Index file")
		return nil
	}

	packageFolder, err := common.GetDefaultPkgFolder()
	if err != nil {
		formatter.PrintErrorMessage("Cannot Parse Board Index file")
		return nil
	}

	bs, err := boards.Find(packageFolder)
	if err != nil {
		formatter.PrintErrorMessage("Cannot Parse Board Index file")
		return nil
	}

	ss := sketches.Find(homeFolder)

	sketch, exists := ss[arduinoBoardAttachFlags.SketchName]
	if !exists {
		formatter.PrintErrorMessage("Cannot find specified sketch in the Sketchbook")
		return nil
	}

	deviceURI, err := url.Parse(arduinoBoardAttachFlags.BoardURI)
	if err != nil {
		formatter.PrintErrorMessage("The provided Device URL is not in a valid format")
		return nil
	}

	var findBoardFunc func(boards.Boards, *discovery.Monitor, *url.URL) *boards.Board
	var Type string

	if validSerialBoardURIRegexp.Match([]byte(arduinoBoardAttachFlags.BoardURI)) {
		findBoardFunc = findSerialConnectedBoard
		Type = "serial"
	} else if validNetworkBoardURIRegexp.Match([]byte(arduinoBoardAttachFlags.BoardURI)) {
		findBoardFunc = findNetworkConnectedBoard
		Type = "network"
	} else {
		formatter.PrintErrorMessage("Invalid device port type provided. Accepted types are: serial://, tty://, http://, https://, tcp://, udp://")
		return nil
	}

	board := findBoardFunc(bs, monitor, deviceURI)

	sketch.Metadata.CPU = sketches.MetadataCPU{
		Fqbn: board.Fqbn,
		Name: board.Name,
		Type: Type,
	}
	err = sketch.ExportMetadata()
	if err != nil {
		formatter.PrintError(err)
	}
	formatter.PrintResult("BOARD ATTACHED")
	return nil
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
		formatter.PrintErrorMessage("No Supported board has been found at the specified board URI")
		return nil
	}

	board := bs.ByVidPid(serialDevice.VendorID, serialDevice.ProductID)
	if board == nil {
		formatter.PrintErrorMessage("No Supported board has been found, try either install new cores or check your board URI")
		return nil
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
		formatter.PrintErrorMessage("No Supported board has been found at the specified board URI, try either install new cores or check your board URI")
		return nil
	}

	formatter.Print("SUPPORTED BOARD FOUND:")
	return bs.ByID(networkDevice.Name)
}
