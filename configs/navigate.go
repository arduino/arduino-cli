package configs

import (
	"fmt"

	paths "github.com/arduino/go-paths-helper"
	homedir "github.com/mitchellh/go-homedir"
)

func Navigate(root, pwd string) Configuration {
	fmt.Println("Navigate", root, pwd)
	home, err := homedir.Dir()
	if err != nil {
		panic(err) // Should never happen
	}

	// Default configuration
	config := Configuration{
		SketchbookDir: paths.New(home, "Arduino"),
		DataDir:       paths.New(home, ".arduino15"),
	}

	// Search for arduino-cli.yaml in current folder
	_ = config.LoadFromYAML(paths.New(pwd, "arduino-cli.yaml"))

	return config
}
