package catalog

import (
	"os"

	"github.com/arduino/arduino-cli/i18n/cmd/ast"
	"github.com/spf13/cobra"
)

var generateCatalogCommand = &cobra.Command{
	Use:   "generate [source files]",
	Short: "generates the en catalog from source files",
	Args:  cobra.MinimumNArgs(1),
	Run:   generateCatalog,
}

func generateCatalog(cmd *cobra.Command, args []string) {
	catalog := ast.GenerateCatalog(args)
	catalog.Write(os.Stdout)
}
