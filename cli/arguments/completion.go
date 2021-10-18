package arguments

import (
	"context"

	"github.com/arduino/arduino-cli/cli/instance"
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
// It returns a list of protocols available
func GetInstalledProtocols(toComplete string) []string {
	inst := instance.CreateAndInit() // TODO optimize this: it does not make sense to create an instance everytime

	detectedBoards, _ := board.List(&rpc.BoardListRequest{
		Instance: inst,
	})
	var res []string
	for _, i := range detectedBoards {
		res = append(res, i.Port.Protocol)
	}
	return res
}
