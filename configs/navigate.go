package configs

import (
	paths "github.com/arduino/go-paths-helper"
	homedir "github.com/mitchellh/go-homedir"
)

func Navigate(root, pwd string) Configuration {
	home, err := homedir.Dir()
	if err != nil {
		panic(err) // Should never happen
	}

	return Configuration{
		SketchbookDir: paths.New(home, "Arduino"),
		DataDir:       paths.New(home, ".arduino15"),
	}
}
