package transifex

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Command is the transifex command
var Command = &cobra.Command{
	Use:              "transifex",
	Short:            "transifex",
	PersistentPreRun: preRun,
}

var project string
var resource string
var apiKey string
var languages = []string{}

func init() {
	Command.AddCommand(pullTransifexCommand)
	Command.AddCommand(pushTransifexCommand)

	Command.PersistentFlags().StringSliceVarP(&languages, "languages", "l", nil, "languages")
	Command.MarkFlagRequired("languages")
}

func preRun(cmd *cobra.Command, args []string) {
	project = os.Getenv("TRANSIFEX_PROJECT")
	resource = os.Getenv("TRANSIFEX_RESOURCE")
	apiKey = os.Getenv("TRANSIFEX_RESOURCE")

	if project = os.Getenv("TRANSIFEX_PROJECT"); project == "" {
		fmt.Println("missing TRANSIFEX_PROJECT environment variable")
		os.Exit(1)
	}

	if resource = os.Getenv("TRANSIFEX_RESOURCE"); resource == "" {
		fmt.Println("missing TRANSIFEX_RESOURCE environment variable")
		os.Exit(1)
	}

	if apiKey = os.Getenv("TRANSIFEX_API_KEY"); apiKey == "" {
		fmt.Println("missing TRANSIFEX_API_KEY environment variable")
		os.Exit(1)
	}

	return
}
