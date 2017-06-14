package libCmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	// LibVersion represents the `arduino lib` package version number.
	LibVersion string = "0.0.1-pre-alpha"
)

func init() {
	LibRoot.AddCommand(libVersionCmd)
}

// LibVersionCmd represents the version command.
var libVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version Number of arduino lib",
	Long:  `Shows version Number of arduino lib which is installed on your system.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("arduino lib V. %s\n", LibsVersion)
	},
}
