// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

package commands

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/pkg/fqbn"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
	"go.bug.st/f"
)

// BoardUploadDetails returns all details for uploading to a board including tools and HW identifiers.
// This command basically gather al the information and translates it into the required grpc struct properties
func (s *arduinoCoreServerImpl) BoardUploadDetails(ctx context.Context, req *rpc.BoardUploadDetailsRequest) (*rpc.BoardUploadDetailsResponse, error) {
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	fqbn, err := fqbn.Parse(req.GetFqbn())
	if err != nil {
		return nil, &cmderrors.InvalidFQBNError{Cause: err}
	}

	// Find target board and board properties
	_, boardPlatform, board, boardProperties, buildPlatform, err := pme.ResolveFQBN(fqbn)
	{
		if boardPlatform == nil {
			return nil, &cmderrors.PlatformNotFoundError{
				Platform: fmt.Sprintf("%s:%s", fqbn.Vendor, fqbn.Architecture),
				Cause:    err,
			}
		} else if err != nil {
			return nil, &cmderrors.UnknownFQBNError{Cause: err}
		}
		logrus.WithField("boardPlatform", boardPlatform).
			WithField("board", board).
			WithField("buildPlatform", buildPlatform).
			Tracef("Upload data")
	}

	// Extract programmer properties (when specified)
	var programmer *cores.Programmer
	{
		if req.GetProgrammer() != "" {
			programmer = boardPlatform.Programmers[req.GetProgrammer()]
			if programmer == nil {
				// Try to find the programmer in the referenced build platform
				programmer = buildPlatform.Programmers[req.GetProgrammer()]
			}
			if programmer == nil {
				return nil, &cmderrors.ProgrammerNotFoundError{Programmer: req.GetProgrammer()}
			}
		}
	}

	// Determine upload action
	// -----------------------
	// Burn bootloader set? | Programmer specified? | Action
	// -------------------- | --------------------- | --------------------
	// false                | false                 | "upload"
	// false                | true                  | "program"
	// true                 | false                 | Error: missing programmer
	// true                 | true                  | "bootloader"
	var action string
	{
		if req.GetBurnBootloader() && req.GetProgrammer() == "" {
			return nil, &cmderrors.MissingProgrammerError{}
		}
		action = "upload"
		if req.GetBurnBootloader() {
			action = "bootloader"
		} else if programmer != nil {
			action = "program"
		}
	}

	// Determine upload recipe
	// -----------------------
	// create a temporary configuration only for the selection of upload recipe
	var uploadToolRecipe string
	var uploadToolPlatform *cores.PlatformRelease
	{
		props := properties.NewMap()
		props.Merge(boardPlatform.Properties)
		props.Merge(boardPlatform.RuntimeProperties())
		props.Merge(boardProperties)
		if programmer != nil {
			props.Merge(programmer.Properties)
		}

		uploadToolRecipe, err = getToolRecipeID(props, action, req.GetProtocol())
		if err != nil {
			return nil, err
		}

		if programmer != nil {
			uploadToolPlatform = programmer.PlatformRelease
		} else {
			uploadToolPlatform = boardPlatform
		}
		logrus.WithField("uploadToolRecipeID", uploadToolRecipe).
			WithField("uploadToolPlatform", uploadToolPlatform).
			Trace("Upload recipe")

		if split := strings.Split(uploadToolRecipe, ":"); len(split) > 2 {
			return nil, &cmderrors.InvalidPlatformPropertyError{
				Property: fmt.Sprintf("%s.tool.%s", action, req.GetProtocol()),
				Value:    uploadToolRecipe}
		} else if len(split) == 2 {
			p := pme.FindPlatform(&packagemanager.PlatformReference{
				Package:              split[0],
				PlatformArchitecture: boardPlatform.Platform.Architecture,
			})
			if p == nil {
				return nil, &cmderrors.PlatformNotFoundError{Platform: split[0] + ":" + boardPlatform.Platform.Architecture}
			}
			uploadToolRecipe = split[1]
			uploadToolPlatform = pme.GetInstalledPlatformRelease(p)
			if uploadToolPlatform == nil {
				return nil, &cmderrors.PlatformNotFoundError{Platform: split[0] + ":" + boardPlatform.Platform.Architecture}
			}
		}
	}

	// Build configuration for upload
	// ------------------------------
	uploadProperties := properties.NewMap()
	{
		uploadProperties.Set("runtime.upload.action", action)
		uploadProperties.Set("runtime.upload.programmer", req.GetProgrammer())
		if uploadToolPlatform != nil {
			uploadProperties.Merge(uploadToolPlatform.Properties)
		}
		uploadProperties.Set("runtime.os", properties.GetOSSuffix())
		uploadProperties.Merge(boardPlatform.Properties)
		uploadProperties.Merge(boardPlatform.RuntimeProperties())
		uploadProperties.Merge(overrideProtocolProperties(action, req.GetProtocol(), boardProperties))
		uploadProperties.Merge(uploadProperties.SubTree("tools." + uploadToolRecipe))
		if programmer != nil {
			uploadProperties.Merge(programmer.Properties)
		}

		// Add user provided custom upload properties
		if p, err := properties.LoadFromSlice(req.GetCustomUploadProperties()); err == nil {
			uploadProperties.Merge(p)
		} else {
			return nil, fmt.Errorf("invalid build properties: %w", err)
		}

		if !uploadProperties.ContainsKey("upload.protocol") && programmer == nil {
			return nil, &cmderrors.ProgrammerRequiredForUploadError{}
		}
	}

	// Determine required tools for upload
	// -----------------------------------
	// Determine the required tools and add metadata for them to the firmware details
	allInstalledTools := pme.GetAllInstalledToolsReleases()
	onlyRequiredVersionTools, err := pme.FindToolsRequiredForBuild(boardPlatform, buildPlatform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.Tr("finding required tools"), err)
	}

	// Extract the tools required for the upload from the upload recipe
	// and add them to the firmware details.
	var uploadTools []*rpc.ToolsDependencies
	{
		// Remove all {runtime.tools.*.path} values from the upload properties.
		// This way we can find the tools required for upload by looking at the
		// {runtime.tools.*.path} variables remaining in the upload pattern after
		// a full variable expansion.
		uploadPropertiesNoRuntime := uploadProperties.Clone()
		for key := range uploadProperties.IterKeys() {
			if strings.HasPrefix(key, "runtime.tools.") && strings.HasSuffix(key, ".path") {
				uploadPropertiesNoRuntime.Remove(key)
			}
		}

		runtimeToolsRegex := regexp.MustCompile(`\{runtime\.tools\.([^}]+)\.path\}`)
		uploadPattern := uploadPropertiesNoRuntime.ExpandPropsInString(uploadPropertiesNoRuntime.Get(action + ".pattern"))
		toolsFound := f.Map(runtimeToolsRegex.FindAllStringSubmatch(uploadPattern, -1), func(match []string) string { return match[1] })

		toolsAdded := map[*cores.ToolRelease]bool{}
		for _, toolID := range toolsFound {
			toolMatchesToolID := func(tool *cores.ToolRelease) bool {
				return tool.Tool.Name == toolID || tool.Tool.Name+"."+tool.Version.String() == toolID
			}
			searchIn := func(l []*cores.ToolRelease) *cores.ToolRelease {
				if idx := slices.IndexFunc(l, toolMatchesToolID); idx != -1 {
					return l[idx]
				}
				return nil
			}
			tool := searchIn(onlyRequiredVersionTools)
			if tool == nil {
				tool = searchIn(allInstalledTools)
			}
			if tool == nil {
				return nil, fmt.Errorf("%s: %w", i18n.Tr("required tool %s not found", toolID), err)
			}
			if toolsAdded[tool] {
				continue
			}
			toolsAdded[tool] = true
			uploadTools = append(uploadTools, tool.ToRpcToolsDependencies())
		}
	}

	return &rpc.BoardUploadDetailsResponse{
		UploadProperties: uploadProperties.CloneAsMap(),
		UploadTools:      uploadTools,
	}, nil
}
