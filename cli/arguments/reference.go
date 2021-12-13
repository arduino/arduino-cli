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

	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/core"
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
	ret.PackageName = toks[0]
	if toks[1] == "" {
		return nil, fmt.Errorf(tr("invalid empty core architecture '%s'"), arg)
	}
	ret.Architecture = toks[1]

	// Now that we have the required informations in `ret` we can
	// try to use core.GetPlatforms to optimize what the user typed
	// (by replacing the PackageName and Architecture in ret with the content of core.GetPlatform())
	platforms, _ := core.GetPlatforms(&rpc.PlatformListRequest{
		Instance:      instance.CreateAndInit(),
		UpdatableOnly: false,
		All:           true, // this is true because we want also the installable platforms
	})
	var found = 0
	for _, platform := range platforms {
		if strings.EqualFold(platform.GetId(), ret.PackageName+":"+ret.Architecture) {
			logrus.Infof("Found possible match for reference %s -> %s", arg, platform.GetId())
			toks = strings.Split(platform.GetId(), ":")
			found++
		}
	}
	// replace the returned Reference only if only one occurrence is found, otherwise leave it as is
	if found == 1 {
		ret.PackageName = toks[0]
		ret.Architecture = toks[1]
	} else {
		logrus.Warnf("Found %d platform for reference: %s", found, arg)
	}
	return ret, nil
}
