package cmd

import (
	"time"

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

var arduinoBoardAttachCmd = &cobra.Command{}

// executeBoardListCommand detects and lists the connected arduino boards
// (either via serial or network ports).
func executeBoardListCommand(cmd *cobra.Command, args []string) {
	monitor := discovery.New(time.Millisecond)
	monitor.Start()

	time.Sleep(time.Second)

	formatter.Print(*monitor)
	//monitor.Stop() //If called will slow like 1sec the program to close after print, with the same result (tested).
}
