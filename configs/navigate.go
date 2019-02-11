package configs

import (
	"path/filepath"
	"strings"

	paths "github.com/arduino/go-paths-helper"
)

func (c *Configuration) Navigate(root, pwd string) {
	relativePath, err := filepath.Rel(root, pwd)
	if err != nil {
		return
	}

	// From the root to the current folder, search for arduino-cli.yaml files
	parts := strings.Split(relativePath, string(filepath.Separator))
	for i := range parts {
		path := paths.New(root)
		path = path.Join(parts[:i+1]...)
		path = path.Join("arduino-cli.yaml")
		_ = c.LoadFromYAML(path)
	}
}
