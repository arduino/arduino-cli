package arguments

import (
	"context"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/core"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// GetInstalledBoards is an helper function useful to autocomplete.
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

// GetInstalledProtocols is an helper function useful to autocomplete.
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

// GetInstalledProgrammers is an helper function useful to autocomplete.
// It returns a list of programmers available based on the installed boards
func GetInstalledProgrammers(toComplete string) []string {
	inst := instance.CreateAndInit() // TODO optimize this: it does not make sense to create an instance everytime
	pm := commands.GetPackageManager(inst.Id)

	// we need the list of the available fqbn in order to get the list of the programmers
	list, _ := board.ListAll(context.Background(), &rpc.BoardListAllRequest{
		Instance:            inst,
		SearchArgs:          nil,
		IncludeHiddenBoards: false,
	})

	// use this strange map because it should be more optimized
	// we use only the key and not the value because we do not need it
	installedProgrammers := make(map[string]struct{})
	for _, i := range list.GetBoards() {
		fqbn, _ := cores.ParseFQBN(i.Fqbn)
		_, boardPlatform, _, _, _, _ := pm.ResolveFQBN(fqbn)
		for programmerID := range boardPlatform.Programmers {
			installedProgrammers[programmerID] = struct{}{}
		}
	}

	res := make([]string, len(installedProgrammers))
	i := 0
	for k := range installedProgrammers {
		res[i] = k
		i++
	}
	return res
}

// GetUninstallableCores is an helper function useful to autocomplete.
// It returns a list of cores which can be uninstalled
func GetUninstallableCores(toComplete string) []string {
	inst := instance.CreateAndInit() // TODO optimize this: it does not make sense to create an instance everytime

	platforms, _ := core.GetPlatforms(&rpc.PlatformListRequest{
		Instance:      inst,
		UpdatableOnly: false,
		All:           false,
	})
	var res []string
	// transform the data structure for the completion
	for _, i := range platforms {
		res = append(res, i.GetId())
	}
	return res
}

// GetInstallableCores is an helper function useful to autocomplete.
// It returns a list of cores which can be installed/downloaded
func GetInstallableCores(toComplete string) []string {
	inst := instance.CreateAndInit() // TODO optimize this: it does not make sense to create an instance everytime

	platforms, _ := core.PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  "",
		AllVersions: false,
	})
	var res []string
	// transform the data structure for the completion
	for _, i := range platforms.SearchOutput {
		res = append(res, i.GetId())
	}
	return res
}
