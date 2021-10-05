package arguments

import (
	"context"

	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/board"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// GetInstalledBoards is an helper function usefull to autocomplete.
// It returns a list of fqbn
// it's taken from cli/board/listall.go
func GetInstalledBoards(toComplete string) []string {
	inst := instance.CreateAndInit() // TODO optimize this: it does not make sense to create an instance everytime

	list, _ := board.ListAll(context.Background(), &rpc.BoardListAllRequest{
		Instance:            inst,
		SearchArgs:          nil,
		IncludeHiddenBoards: false,
	})
	var res []string
	// transform the data structure for the completion
	for _, i := range list.GetBoards() {
		res = append(res, i.Fqbn)
	}
	return res
}

// GetInstalledProtocols is an helper function usefull to autocomplete.
// It returns a list of protocols available based on the installed boards
func GetInstalledProtocols(toComplete string) []string {
	inst := instance.CreateAndInit() // TODO optimize this: it does not make sense to create an instance everytime
	pm := commands.GetPackageManager(inst.Id)
	boards := pm.InstalledBoards()

	// use this strange map because it should be more optimized
	// we use only the key and not the value because we do not need it
	intalledProtocolsMap := make(map[string]struct{})
	for _, board := range boards {
		// we filter and elaborate a bit the informations present in Properties
		for _, protocol := range board.Properties.SubTree("upload.tool").FirstLevelKeys() {
			if protocol != "default" { // remove this value since it's the default one
				intalledProtocolsMap[protocol] = struct{}{}
			}
		}
	}
	res := make([]string, len(intalledProtocolsMap))
	i := 0
	for k := range intalledProtocolsMap {
		res[i] = k
		i++
	}
	return res
}
