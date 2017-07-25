package cmd

import (
	"github.com/spf13/cobra/doc"
)

// generateManPages generates man pages for all commands and puts them in $PROJECT_DIR/manpages
func generateManPages() error {
	header := &doc.GenManHeader{
		Title:   "ARDUINO COMMAND LINE MANUAL",
		Section: "1",
	}
	//out := new(bytes.Buffer)
	//doc.GenMan(cmd.RootCmd, header, out)
	return doc.GenManTree(arduinoCmd, header, "./docs/manpages")
}
