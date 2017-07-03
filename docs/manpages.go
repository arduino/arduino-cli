package docs

import (
	"github.com/bcmi-labs/arduino-cli/cmd"
	"github.com/spf13/cobra/doc"
)

// GenerateManPages generates man pages for all commands and puts them in $PROJECT_DIR/manpages
func GenerateManPages() {
	header := &doc.GenManHeader{
		Title:   "ARDUINO COMMAND LINE MANUAL",
		Section: "1",
	}
	//out := new(bytes.Buffer)
	//doc.GenMan(cmd.RootCmd, header, out)
	//logrus.Info(out.String())
	doc.GenManTree(cmd.RootCmd, header, "./docs/manpages")
}
