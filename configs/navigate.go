package configs

import (
	"path/filepath"
	"strings"

	paths "github.com/arduino/go-paths-helper"
	homedir "github.com/mitchellh/go-homedir"
)

func Navigate(root, pwd string) Configuration {
	relativePath, err := filepath.Rel(root, pwd)
	if err != nil {
		panic(err)
	}

	home, err := homedir.Dir()
	if err != nil {
		panic(err) // Should never happen
	}

	// Default configuration
	config := Configuration{
		SketchbookDir: paths.New(home, "Arduino"),
		DataDir:       paths.New(home, ".arduino15"),
	}

	// From the root to the current folder, search for arduino-cli.yaml files
	parts := strings.Split(relativePath, string(filepath.Separator))
	for i := range parts {
		path := paths.New(root)
		path = path.Join(parts[:i+1]...)
		path = path.Join("arduino-cli.yaml")
		_ = config.LoadFromYAML(path)
	}

	return config
}
