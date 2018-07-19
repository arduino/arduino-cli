package board

import (
	"sort"

	"github.com/bcmi-labs/arduino-cli/commands"

	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"

	"github.com/spf13/cobra"
)

func initListAllCommand() *cobra.Command {
	listAllCommand := &cobra.Command{
		Use:     "listall",
		Short:   "List all known boards.",
		Long:    "List all boards that have the support platform installed.",
		Example: "arduino board listall",
		Args:    cobra.NoArgs,
		Run:     runListAllCommand,
	}
	return listAllCommand
}

// runListAllCommand list all installed boards
func runListAllCommand(cmd *cobra.Command, args []string) {
	pm := commands.InitPackageManager()

	list := &output.BoardList{}
	for _, targetPackage := range pm.GetPackages().Packages {
		for _, platform := range targetPackage.Platforms {
			platformRelease := platform.GetInstalled()
			if platformRelease == nil {
				continue
			}
			for _, board := range platformRelease.Boards {
				list.Boards = append(list.Boards, &output.BoardListItem{
					Name: board.Name(),
					Fqbn: board.FQBN(),
				})
			}
		}
	}
	sort.Sort(list)
	formatter.Print(list)
}
