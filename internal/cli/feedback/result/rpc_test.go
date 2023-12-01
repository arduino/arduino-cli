// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package result_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/stretchr/testify/require"
)

func getStructJsonTags(t *testing.T, a any) []string {
	tags := []string{}
	rt := reflect.TypeOf(a)
	if rt.Kind() != reflect.Struct {
		rt = rt.Elem()
		require.Equal(t, reflect.Struct, rt.Kind())
	}
	for i := 0; i < rt.NumField(); i++ {
		tag := rt.Field(i).Tag.Get("json")
		if tag == "" {
			continue
		}
		key, _, _ := strings.Cut(tag, ",")
		tags = append(tags, key)
	}
	return tags
}

func mustContainsAllPropertyOfRpcStruct(t *testing.T, a, b any, excludeFields ...string) {
	// must not be the same pointer, a and b struct must be of different type
	require.NotSame(t, a, b)
	rta, rtb := reflect.TypeOf(a), reflect.TypeOf(b)
	if rta.Kind() != reflect.Struct {
		rta = rta.Elem()
		require.Equal(t, reflect.Struct, rta.Kind())
	}
	if rtb.Kind() != reflect.Struct {
		rtb = rtb.Elem()
		require.Equal(t, reflect.Struct, rtb.Kind())
	}
	require.NotEqual(t, rta.String(), rtb.String())

	aTags := getStructJsonTags(t, a)
	bTags := getStructJsonTags(t, b)
	if len(excludeFields) > 0 {
		aTags = slices.DeleteFunc(aTags, func(s string) bool { return slices.Contains(excludeFields, s) })
		bTags = slices.DeleteFunc(bTags, func(s string) bool { return slices.Contains(excludeFields, s) })
	}
	require.ElementsMatch(t, aTags, bTags)
}

func TestAllFieldAreMapped(t *testing.T) {
	// Our PlatformSummary expands the PlatformMetadata without the need to nest it as the rpc does.
	platformSummaryRpc := &rpc.PlatformSummary{
		InstalledVersion: "1.0.0",
		LatestVersion:    "1.0.0",
	}
	platformSummaryRpcTags := getStructJsonTags(t, platformSummaryRpc)
	platformSummaryRpcTags = append(platformSummaryRpcTags, getStructJsonTags(t, platformSummaryRpc.GetMetadata())...)
	platformSummaryRpcTags = slices.DeleteFunc(platformSummaryRpcTags, func(s string) bool { return s == "metadata" })

	platformSummaryResult := result.NewPlatformSummary(platformSummaryRpc)
	platformSummaryResultTags := getStructJsonTags(t, platformSummaryResult)

	require.ElementsMatch(t, platformSummaryRpcTags, platformSummaryResultTags)

	platformRelease := &rpc.PlatformRelease{}
	platformReleaseResult := result.NewPlatformRelease(platformRelease)
	mustContainsAllPropertyOfRpcStruct(t, platformRelease, platformReleaseResult)

	libraryRpc := &rpc.Library{}
	libraryResult := result.NewLibrary(libraryRpc)
	mustContainsAllPropertyOfRpcStruct(t, libraryRpc, libraryResult)

	libraryReleaseRpc := &rpc.LibraryRelease{}
	libraryReleaseResult := result.NewLibraryRelease(libraryReleaseRpc)
	mustContainsAllPropertyOfRpcStruct(t, libraryReleaseRpc, libraryReleaseResult)

	installedLibrary := &rpc.InstalledLibrary{}
	installedLibraryResult := result.NewInstalledLibrary(installedLibrary)
	mustContainsAllPropertyOfRpcStruct(t, installedLibrary, installedLibraryResult)

	downloadResource := &rpc.DownloadResource{}
	downloadResourceResult := result.NewDownloadResource(downloadResource)
	mustContainsAllPropertyOfRpcStruct(t, downloadResource, downloadResourceResult)

	libraryDependencyRpc := &rpc.LibraryDependency{}
	libraryDependencyResult := result.NewLibraryDependency(libraryDependencyRpc)
	mustContainsAllPropertyOfRpcStruct(t, libraryDependencyRpc, libraryDependencyResult)

	portRpc := &rpc.Port{}
	portResult := result.NewPort(portRpc)
	mustContainsAllPropertyOfRpcStruct(t, portRpc, portResult)

	boardDetailsResponseRpc := &rpc.BoardDetailsResponse{}
	boardDetailsResponseResult := result.NewBoardDetailsResponse(boardDetailsResponseRpc)
	mustContainsAllPropertyOfRpcStruct(t, boardDetailsResponseRpc, boardDetailsResponseResult)

	packageRpc := &rpc.Package{}
	packageResult := result.NewPackage(packageRpc)
	mustContainsAllPropertyOfRpcStruct(t, packageRpc, packageResult)

	helpRpc := &rpc.Help{}
	helpResult := result.NewHelp(helpRpc)
	mustContainsAllPropertyOfRpcStruct(t, helpRpc, helpResult)

	boardPlatformRpc := &rpc.BoardPlatform{}
	boardPlatformResult := result.NewBoardPlatform(boardPlatformRpc)
	mustContainsAllPropertyOfRpcStruct(t, boardPlatformRpc, boardPlatformResult)

	toolsDependencyRpc := &rpc.ToolsDependencies{}
	toolsDependencyResult := result.NewToolsDependency(toolsDependencyRpc)
	mustContainsAllPropertyOfRpcStruct(t, toolsDependencyRpc, toolsDependencyResult)

	systemRpc := &rpc.Systems{}
	systemResult := result.NewSystem(systemRpc)
	mustContainsAllPropertyOfRpcStruct(t, systemRpc, systemResult)

	configOptionRpc := &rpc.ConfigOption{}
	configOptionResult := result.NewConfigOption(configOptionRpc)
	mustContainsAllPropertyOfRpcStruct(t, configOptionRpc, configOptionResult)

	configValueRpc := &rpc.ConfigValue{}
	configValueResult := result.NewConfigValue(configValueRpc)
	mustContainsAllPropertyOfRpcStruct(t, configValueRpc, configValueResult)

	programmerRpc := &rpc.Programmer{}
	programmerResult := result.NewProgrammer(programmerRpc)
	mustContainsAllPropertyOfRpcStruct(t, programmerRpc, programmerResult)

	boardIdentificationPropertiesRpc := &rpc.BoardIdentificationProperties{}
	boardIdentificationPropertiesResult := result.NewBoardIndentificationProperty(boardIdentificationPropertiesRpc)
	mustContainsAllPropertyOfRpcStruct(t, boardIdentificationPropertiesRpc, boardIdentificationPropertiesResult)

	boardListAllResponseRpc := &rpc.BoardListAllResponse{}
	boardListAllResponseResult := result.NewBoardListAllResponse(boardListAllResponseRpc)
	mustContainsAllPropertyOfRpcStruct(t, boardListAllResponseRpc, boardListAllResponseResult)

	boardListItemRpc := &rpc.BoardListItem{}
	boardListItemResult := result.NewBoardListItem(boardListItemRpc)
	mustContainsAllPropertyOfRpcStruct(t, boardListItemRpc, boardListItemResult)

	platformRpc := &rpc.Platform{}
	platformResult := result.NewPlatform(platformRpc)
	mustContainsAllPropertyOfRpcStruct(t, platformRpc, platformResult)

	platformMetadataRpc := &rpc.PlatformMetadata{}
	platformMetadataResult := result.NewPlatformMetadata(platformMetadataRpc)
	mustContainsAllPropertyOfRpcStruct(t, platformMetadataRpc, platformMetadataResult)

	detectedPortRpc := &rpc.DetectedPort{}
	detectedPortResult := result.NewDetectedPort(detectedPortRpc)
	mustContainsAllPropertyOfRpcStruct(t, detectedPortRpc, detectedPortResult)

	libraryResolveDependenciesResponseRpc := &rpc.LibraryResolveDependenciesResponse{}
	libraryResolveDependenciesResponseResult := result.NewLibraryResolveDependenciesResponse(libraryResolveDependenciesResponseRpc)
	mustContainsAllPropertyOfRpcStruct(t, libraryResolveDependenciesResponseRpc, libraryResolveDependenciesResponseResult)

	libraryDependencyStatusRpc := &rpc.LibraryDependencyStatus{}
	libraryDependencyStatusResult := result.NewLibraryDependencyStatus(libraryDependencyStatusRpc)
	mustContainsAllPropertyOfRpcStruct(t, libraryDependencyStatusRpc, libraryDependencyStatusResult)

	librarySearchResponseRpc := &rpc.LibrarySearchResponse{}
	librarySearchResponseResult := result.NewLibrarySearchResponse(librarySearchResponseRpc)
	mustContainsAllPropertyOfRpcStruct(t, librarySearchResponseRpc, librarySearchResponseResult)

	searchedLibraryRpc := &rpc.SearchedLibrary{}
	searchedLibraryResult := result.NewSearchedLibrary(searchedLibraryRpc)
	mustContainsAllPropertyOfRpcStruct(t, searchedLibraryRpc, searchedLibraryResult)

	monitorPortSettingDescriptorRpc := &rpc.MonitorPortSettingDescriptor{}
	monitorPortSettingDescriptorResult := result.NewMonitorPortSettingDescriptor(monitorPortSettingDescriptorRpc)
	mustContainsAllPropertyOfRpcStruct(t, monitorPortSettingDescriptorRpc, monitorPortSettingDescriptorResult)

	compileResponseRpc := &rpc.CompileResponse{}
	compileResponseResult := result.NewCompileResponse(compileResponseRpc)
	mustContainsAllPropertyOfRpcStruct(t, compileResponseRpc, compileResponseResult, "progress")

	executableSectionSizeRpc := &rpc.ExecutableSectionSize{}
	executableSectionSizeResult := result.NewExecutableSectionSize(executableSectionSizeRpc)
	mustContainsAllPropertyOfRpcStruct(t, executableSectionSizeRpc, executableSectionSizeResult)

	installedPlatformReferenceRpc := &rpc.InstalledPlatformReference{}
	installedPlatformReferenceResult := result.NewInstalledPlatformReference(installedPlatformReferenceRpc)
	mustContainsAllPropertyOfRpcStruct(t, installedPlatformReferenceRpc, installedPlatformReferenceResult)

	boardListWatchResponseRpc := &rpc.BoardListWatchResponse{}
	boardListWatchResponseResult := result.NewBoardListWatchResponse(boardListWatchResponseRpc)
	mustContainsAllPropertyOfRpcStruct(t, boardListWatchResponseRpc, boardListWatchResponseResult)

	compileDiagnosticRpc := &rpc.CompileDiagnostic{}
	compileDiagnosticResult := result.NewCompileDiagnostic(compileDiagnosticRpc)
	mustContainsAllPropertyOfRpcStruct(t, compileDiagnosticRpc, compileDiagnosticResult)

	compileDiagnosticContextRpc := &rpc.CompileDiagnosticContext{}
	compileDiagnosticContextResult := result.NewCompileDiagnosticContext(compileDiagnosticContextRpc)
	mustContainsAllPropertyOfRpcStruct(t, compileDiagnosticContextRpc, compileDiagnosticContextResult)

	compileDiagnosticNoteRpc := &rpc.CompileDiagnosticNote{}
	compileDiagnosticNoteResult := result.NewCompileDiagnosticNote(compileDiagnosticNoteRpc)
	mustContainsAllPropertyOfRpcStruct(t, compileDiagnosticNoteRpc, compileDiagnosticNoteResult)

	isDebugSupportedResponseRpc := &rpc.IsDebugSupportedResponse{}
	isDebugSupportedResponseResult := result.NewIsDebugSupportedResponse(isDebugSupportedResponseRpc)
	mustContainsAllPropertyOfRpcStruct(t, isDebugSupportedResponseRpc, isDebugSupportedResponseResult)
}

func TestEnumsMapsEveryRpcCounterpart(t *testing.T) {
	t.Run("LibraryLocation enums maps every element", func(t *testing.T) {
		results := make([]result.LibraryLocation, 0, len(rpc.LibraryLocation_name))
		for key := range rpc.LibraryLocation_name {
			results = append(results, result.NewLibraryLocation(rpc.LibraryLocation(key)))
		}
		require.NotEmpty(t, results)
		require.Len(t, results, len(rpc.LibraryLocation_name))
		require.True(t, isUnique(results))
	})
	t.Run("LibraryLayout enums maps every element", func(t *testing.T) {
		results := make([]result.LibraryLayout, 0, len(rpc.LibraryLayout_name))
		for key := range rpc.LibraryLayout_name {
			results = append(results, result.NewLibraryLayout(rpc.LibraryLayout(key)))
		}
		require.NotEmpty(t, results)
		require.Len(t, results, len(rpc.LibraryLayout_name))
		require.True(t, isUnique(results))
	})
	t.Run("LibrarySearchStatus enums maps every element", func(t *testing.T) {
		results := make([]result.LibrarySearchStatus, 0, len(rpc.LibrarySearchStatus_name))
		for key := range rpc.LibrarySearchStatus_name {
			results = append(results, result.NewLibrarySearchStatus(rpc.LibrarySearchStatus(key)))
		}
		require.NotEmpty(t, results)
		require.Len(t, results, len(rpc.LibrarySearchStatus_name))
		require.True(t, isUnique(results))
	})
}

func isUnique[T comparable](s []T) bool {
	seen := map[T]bool{}
	for _, v := range s {
		if _, ok := seen[v]; ok {
			return false
		}
		seen[v] = true
	}
	return true
}
