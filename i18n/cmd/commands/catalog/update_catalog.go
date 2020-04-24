package catalog

import (
	"fmt"
	"os"
	"path"

	"github.com/arduino/arduino-cli/i18n/cmd/po"
	"github.com/spf13/cobra"
)

var updateCatalogCommand = &cobra.Command{
	Use:   "update -l pt_BR [catalog folder]",
	Short: "updates the language catalogs base on changes to the en catalog",
	Args:  cobra.ExactArgs(1),
	Run:   updateCatalog,
}

var languages = []string{}

func init() {
	updateCatalogCommand.Flags().StringSliceVarP(&languages, "languages", "l", nil, "languages")
	updateCatalogCommand.MarkFlagRequired("languages")
}

func updateCatalog(cmd *cobra.Command, args []string) {
	folder := args[0]
	enCatalog := po.Parse(path.Join(folder, "en.po"))

	for _, lang := range languages {
		filename := path.Join(folder, fmt.Sprintf("%s.po", lang))
		langCatalog := po.Parse(filename)

		mergedCatalog := po.Merge(enCatalog, langCatalog)

		os.Remove(filename)
		file, err := os.OpenFile(path.Join(folder, fmt.Sprintf("%s.po", lang)), os.O_CREATE|os.O_RDWR, 0644)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		mergedCatalog.Write(file)
	}
}
