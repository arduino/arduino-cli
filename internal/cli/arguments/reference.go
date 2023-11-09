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

package arguments

import (
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
)

// Reference represents a reference item (core or library) passed to the CLI
// interface
type Reference struct {
	PackageName  string
	Architecture string
	Version      string
}

func (r *Reference) String() string {
	if r.Version != "" {
		return r.PackageName + ":" + r.Architecture + "@" + r.Version
	}
	return r.PackageName + ":" + r.Architecture
}

// ParseReferences is a convenient wrapper that operates on a slice of strings and
// calls ParseReference for each of them. It returns at the first invalid argument.
func ParseReferences(args []string) ([]*Reference, error) {
	ret := []*Reference{}
	for _, arg := range args {
		reference, err := ParseReference(arg)
		if err != nil {
			return nil, err
		}
		ret = append(ret, reference)
	}
	return ret, nil
}

// ParseReference parses a string and returns a Reference object.
// It tries to infer the platform the user is asking for.
// To achieve that, it tries to use github.com/arduino/arduino-cli/commands/core.GetPlatform
// Note that the Reference is returned rightaway if the arg inserted by the user matches perfectly one in the response of core.GetPlatform
// A MultiplePlatformsError is returned if the platform searched by the user matches multiple platforms
func ParseReference(arg string) (*Reference, error) {
	logrus.Infof("Parsing reference %s", arg)
	ret := &Reference{}
	if arg == "" {
		return nil, fmt.Errorf(tr("invalid empty core argument"))
	}

	toks := strings.SplitN(arg, "@", 2)
	if toks[0] == "" {
		return nil, fmt.Errorf(tr("invalid empty core reference '%s'"), arg)
	}
	ret.PackageName = toks[0]
	if len(toks) > 1 {
		if toks[1] == "" {
			return nil, fmt.Errorf(tr("invalid empty core version: '%s'"), arg)
		}
		ret.Version = toks[1]
	}

	toks = strings.Split(ret.PackageName, ":")
	if len(toks) != 2 {
		return nil, fmt.Errorf(tr("invalid item %s"), arg)
	}
	if toks[0] == "" {
		return nil, fmt.Errorf(tr("invalid empty core name '%s'"), arg)
	}
	if toks[1] == "" {
		return nil, fmt.Errorf(tr("invalid empty core architecture '%s'"), arg)
	}
	ret.PackageName = toks[0]
	ret.Architecture = toks[1]

	// Now that we have the required informations in `ret` we can
	// try to use core.PlatformList to optimize what the user typed
	// (by replacing the PackageName and Architecture in ret with the content of core.GetPlatform())
	platforms, _ := core.PlatformSearch(&rpc.PlatformSearchRequest{
		Instance: instance.CreateAndInit(),
	})
	foundPlatforms := []string{}
	for _, platform := range platforms.GetSearchOutput() {
		platformID := platform.GetMetadata().GetId()
		platformUser := ret.PackageName + ":" + ret.Architecture
		// At first we check if the platform the user is searching for matches an available one,
		// this way we do not need to adapt the casing and we can return it directly
		if platformUser == platformID {
			return ret, nil
		}
		if strings.EqualFold(platformUser, platformID) {
			logrus.Infof("Found possible match for reference %s -> %s", platformUser, platformID)
			foundPlatforms = append(foundPlatforms, platformID)
		}
	}
	// replace the returned Reference only if only one occurrence is found,
	// otherwise return an error to the user because we don't know on which platform operate
	if len(foundPlatforms) == 0 {
		return nil, &arduino.PlatformNotFoundError{Platform: arg}
	}
	if len(foundPlatforms) > 1 {
		return nil, &arduino.MultiplePlatformsError{Platforms: foundPlatforms, UserPlatform: arg}
	}
	toks = strings.Split(foundPlatforms[0], ":")
	ret.PackageName = toks[0]
	ret.Architecture = toks[1]
	return ret, nil
}
