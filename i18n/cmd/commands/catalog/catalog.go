package catalog

import "github.com/spf13/cobra"

var Command = &cobra.Command{
	Use:   "catalog",
	Short: "catalog",
}

func init() {
	Command.AddCommand(generateCatalogCommand)
	Command.AddCommand(updateCatalogCommand)
}
