package commands

import (
	"github.com/arduino/arduino-cli/i18n/cmd/commands/catalog"
	"github.com/spf13/cobra"
)

var i18nCommand = &cobra.Command{
	Use:   "i18n",
	Short: "i18n",
}

func init() {
	i18nCommand.AddCommand(catalog.Command)
}

// Execute executes the i18n command
func Execute() error {
	return i18nCommand.Execute()
}
