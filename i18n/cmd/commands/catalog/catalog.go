package catalog

import "github.com/spf13/cobra"

// Command is the catalog command
var Command = &cobra.Command{
	Use:   "catalog",
	Short: "catalog",
}

func init() {
	Command.AddCommand(generateCatalogCommand)
	Command.AddCommand(updateCatalogCommand)
}
