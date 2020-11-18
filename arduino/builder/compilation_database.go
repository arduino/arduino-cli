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
	"os/exec"

	"github.com/arduino/go-paths-helper"
)

type compilationCommand struct {
	Directory string   `json:"directory"`
	Arguments []string `json:"arguments"`
	File      string   `json:"file"`
}

type CompilationDatabase struct {
	contents []compilationCommand
	filename *paths.Path
}

func NewCompilationDatabase(filename *paths.Path) *CompilationDatabase {
	return &CompilationDatabase{
		filename: filename,
	}
}

func (db *CompilationDatabase) UpdateFile(complete bool) {
	// TODO: Read any existing file and use its contents for any
	// kept files, or any files not in db.contents if !complete.
	if jsonContents, err := json.MarshalIndent(db.contents, "", " "); err != nil {
		fmt.Printf("Error serializing compilation database: %s", err)
		return
	} else if err := db.filename.WriteFile(jsonContents); err != nil {
		fmt.Printf("Error writing compilation database: %s", err)
	}
}

func (db *CompilationDatabase) dirForCommand(command *exec.Cmd) string {
	// This mimics what Cmd.Run also does: Use Dir if specified,
	// current directory otherwise
	if command.Dir != "" {
		return command.Dir
	} else {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory for compilation database: %s", err)
			return ""
		}
		return dir
	}
}

func (db *CompilationDatabase) ReplaceEntry(filename *paths.Path, command *exec.Cmd) {
	entry := compilationCommand{
		Directory: db.dirForCommand(command),
		Arguments: command.Args,
		File:      filename.String(),
	}

	db.contents = append(db.contents, entry)
}

func (db *CompilationDatabase) KeepEntry(filename *paths.Path) {
	// TODO
}
