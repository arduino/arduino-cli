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
	"net/url"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/go-paths-helper"
)

type BoardsRegistry struct {
	Boards        []*RegisteredBoard
	fqbnToBoard   map[string]*RegisteredBoard
	aliastToBoard map[string]*RegisteredBoard
}

type RegisteredBoard struct {
	FQBN                *cores.FQBN
	Alias               string
	Name                string
	ExternalPlatformURL *url.URL
}

func NewBoardRegistry() *BoardsRegistry {
	return &BoardsRegistry{
		Boards:        []*RegisteredBoard{},
		fqbnToBoard:   map[string]*RegisteredBoard{},
		aliastToBoard: map[string]*RegisteredBoard{},
	}
}

func (r *BoardsRegistry) addBoard(board *RegisteredBoard) {
	r.Boards = append(r.Boards, board)
	r.fqbnToBoard[board.FQBN.String()] = board
	r.aliastToBoard[board.Alias] = board
}

func (r *BoardsRegistry) FindBoard(fqbnOrAlias string) (*cores.FQBN, *RegisteredBoard, error) {
	if found, ok := r.aliastToBoard[fqbnOrAlias]; ok {
		return found.FQBN, found, nil
	}
	fqbn, err := cores.ParseFQBN(fqbnOrAlias)
	if err != nil {
		return nil, nil, err
	}
	if found, ok := r.fqbnToBoard[fqbn.StringWithoutConfig()]; ok {
		return fqbn, found, nil
	}
	return fqbn, nil, nil
}

func (r *BoardsRegistry) SearchBoards(query string) []*RegisteredBoard {
	found := []*RegisteredBoard{}
	contains := func(a string, b string) bool {
		return strings.Contains(strings.ToLower(a), strings.ToLower(b))
	}
	for _, board := range r.Boards {
		if contains(board.Name, query) || contains(board.Alias, query) {
			found = append(found, board)
		}
	}
	return found
}

func LoadBoardRegistry(file *paths.Path) (*BoardsRegistry, error) {

	// TODO...

	fake := NewBoardRegistry()
	fake.addBoard(&RegisteredBoard{
		Name:  "Arduino Uno",
		FQBN:  cores.MustParseFQBN("arduino:avr:uno"),
		Alias: "uno",
	})
	fake.addBoard(&RegisteredBoard{
		Name:  "Arduino Zero",
		FQBN:  cores.MustParseFQBN("arduino:samd:arduino_zero_edbg"),
		Alias: "zero",
	})
	return fake, nil
}
