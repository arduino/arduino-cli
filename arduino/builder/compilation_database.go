package builder

import (
	"encoding/json"
	"fmt"
	"github.com/arduino/go-paths-helper"
	"io/ioutil"
	"os"
	"os/exec"
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
	fmt.Printf("Writing compilation database to \"%s\"...\n", db.filename.String())

	contents := db.contents
	jsonContents, err := json.MarshalIndent(contents, "", " ")
	if err != nil {
		fmt.Printf("Error serializing compilation database: %s", err)
		return
	}
	err = ioutil.WriteFile(db.filename.String(), jsonContents, 0644)
	if err != nil {
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
