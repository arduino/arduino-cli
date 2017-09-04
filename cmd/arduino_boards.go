package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/bcmi-labs/arduino-modules/sketches"

	"github.com/bcmi-labs/arduino-cli/common"

	"github.com/bcmi-labs/arduino-modules/boards"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"

	"github.com/arduino/board-discovery"
	"github.com/spf13/cobra"
)

var arduinoBoardCmd = &cobra.Command{
	Use:   "board",
	Short: `Arduino board commands`,
	Long:  `Arduino board commands`,
	Example: `arduino board list                     # Lists all connected boards
arduino board attach --board 754393135373516091F1 \
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
	Example: `arduino board attach --board 754393135373516091F1 \
					 --sketch mySketch # Attaches a sketch to a board`,
	RunE: executeBoardAttachCommand,
}

// executeBoardListCommand detects and lists the connected arduino boards
// (either via serial or network ports).
func executeBoardListCommand(cmd *cobra.Command, args []string) {
	monitor := discovery.New(time.Millisecond)
	monitor.Start()

	time.Sleep(time.Second)

	formatter.Print(*monitor)
	//monitor.Stop() //If called will slow like 1sec the program to close after print, with the same result (tested).
	// it closes ungracefully, but at the end of the command we can't have races.
}

func executeBoardAttachCommand(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return errors.New("Not accepting additional arguments")
	}

	monitor := discovery.New(time.Millisecond)
	monitor.Start()

	time.Sleep(time.Second)

	formatter.Print(*monitor)

	packageFolder, err := common.GetDefaultPkgFolder()
	if err != nil {
		formatter.PrintErrorMessage("Cannot Parse Board Index file")
		return nil
	}

	homeFolder, err := common.GetDefaultArduinoHomeFolder()
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
	sketch, exists := ss[flag]
	if !exists {
		formatter.PrintErrorMessage("Cannot find specified sketch")
		return nil
	}
	fmt.Println("SUPPORTED BOARDS:")
	for _, device := range monitor.Serial() {
		board := bs.ByVidPid(device.VendorID, device.ProductID)
		if board != nil {
			fmt.Println("NOT FOUND")
		} else {
			fmt.Println(" -", bs.ByVidPid(device.VendorID, device.ProductID).Fqbn)
		}

	}

	return nil
}
