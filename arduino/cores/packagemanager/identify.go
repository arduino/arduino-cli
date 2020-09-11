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

package packagemanager

import (
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	properties "github.com/arduino/go-properties-orderedmap"
)

// IdentifyBoard returns a list of boards matching the provided identification properties.
func (pm *PackageManager) IdentifyBoard(idProps *properties.Map) []*cores.Board {
	if idProps.Size() == 0 {
		return []*cores.Board{}
	}

	checkSuffix := func(props *properties.Map, s string) (present bool, matched bool) {
		for k, v1 := range idProps.AsMap() {
			v2, ok := props.GetOk(k + s)
			if !ok {
				return false, false
			}
			if !strings.EqualFold(v1, v2) {
				return true, false
			}
		}
		return false, true
	}

	foundBoards := []*cores.Board{}
	for _, board := range pm.InstalledBoards() {
		if _, matched := checkSuffix(board.Properties, ""); matched {
			foundBoards = append(foundBoards, board)
			continue
		}
		id := 0
		for {
			present, matched := checkSuffix(board.Properties, fmt.Sprintf(".%d", id))
			if matched {
				foundBoards = append(foundBoards, board)
			}
			if !present && id > 0 { // Always check id 0 and 1 (https://github.com/arduino/arduino-cli/issues/456)
				break
			}
			id++
		}
	}
	return foundBoards
}
