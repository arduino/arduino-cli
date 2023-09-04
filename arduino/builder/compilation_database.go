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

package builder

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-paths-helper"
)

// CompilationDatabase keeps track of all the compile commands run by the builder
type CompilationDatabase struct {
	Contents []CompilationCommand
	File     *paths.Path
}

// CompilationCommand keeps track of a single run of a compile command
type CompilationCommand struct {
	Directory string   `json:"directory"`
	Command   string   `json:"command,omitempty"`
	Arguments []string `json:"arguments,omitempty"`
	File      string   `json:"file"`
}

// NewCompilationDatabase creates an empty CompilationDatabase
func NewCompilationDatabase(filename *paths.Path) *CompilationDatabase {
	return &CompilationDatabase{
		File:     filename,
		Contents: []CompilationCommand{},
	}
}

// LoadCompilationDatabase reads a compilation database from a file
func LoadCompilationDatabase(file *paths.Path) (*CompilationDatabase, error) {
	f, err := file.ReadFile()
	if err != nil {
		return nil, err
	}
	res := NewCompilationDatabase(file)
	return res, json.Unmarshal(f, &res.Contents)
}

// SaveToFile save the CompilationDatabase to file as a clangd-compatible compile_commands.json,
// see https://clang.llvm.org/docs/JSONCompilationDatabase.html
func (db *CompilationDatabase) SaveToFile() {
	if jsonContents, err := json.MarshalIndent(db.Contents, "", " "); err != nil {
		fmt.Println(tr("Error serializing compilation database: %s", err))
		return
	} else if err := db.File.WriteFile(jsonContents); err != nil {
		fmt.Println(tr("Error writing compilation database: %s", err))
	}
}

// Add adds a new CompilationDatabase entry
func (db *CompilationDatabase) Add(target *paths.Path, command *executils.Process) {
	commandDir := command.GetDir()
	if commandDir == "" {
		// This mimics what Cmd.Run also does: Use Dir if specified,
		// current directory otherwise
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(tr("Error getting current directory for compilation database: %s", err))
		}
		commandDir = dir
	}

	entry := CompilationCommand{
		Directory: commandDir,
		Arguments: command.GetArgs(),
		File:      target.String(),
	}

	db.Contents = append(db.Contents, entry)
}
