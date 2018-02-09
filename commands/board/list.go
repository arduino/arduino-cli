package board

import (
	"fmt"
	"os"
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
	Use:     "list",
	Short:   "List connected boards.",
	Long:    "Detects and displays a list of connected boards to the current computer.",
	Example: "arduino board list --timeout 10s",
	Args:    cobra.NoArgs,
	Run:     runListCommand,
}

// runListCommand detects and lists the connected arduino boards
// (either via serial or network ports).
func runListCommand(cmd *cobra.Command, args []string) {
	packageFolder, err := common.PackagesFolder.Get()
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
