package configs

import (
	"fmt"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/bcmi-labs/arduino-cli/pathutils"
)

// ArduinoDataFolder represents the current root of the arduino tree (defaulted to `$HOME/.arduino15` on linux).
var ArduinoDataFolder = pathutils.NewPath("Arduino Data", getDefaultArduinoDataFolder, true)

// SketchbookFolder represents the current root of the sketchbooks tree (defaulted to `$HOME/Arduino`).
var SketchbookFolder = pathutils.NewPath("Sketchbook", getDefaultSketchbookFolder, true)

// TODO: Remove either ArduinoHomeFolder or SketchbookFolder
var ArduinoHomeFolder = SketchbookFolder

// LibrariesFolder is the default folder of downloaded libraries.
var LibrariesFolder = pathutils.NewSubPath("libraries", ArduinoHomeFolder, "libraries", true)

// PackagesFolder is the default folder of downloaded packages.
var PackagesFolder = pathutils.NewSubPath("packages", ArduinoDataFolder, "packages", true)

// CoresFolder gets the default folder of downloaded cores.
func CoresFolder(packageName string) pathutils.Path {
	// TODO: wrong, this is not the correct location of the cores (in Java IDE)
	return pathutils.NewSubPath("cores", PackagesFolder, filepath.Join(packageName, "hardware"), true)
}

// ToolsFolder gets the default folder of downloaded packages.
func ToolsFolder(packageName string) pathutils.Path {
	return pathutils.NewSubPath("tools", PackagesFolder, filepath.Join(packageName, "tools"), true)
}

// DownloadCacheFolder gets a generic cache folder for downloads.
func DownloadCacheFolder(item string) pathutils.Path {
	return pathutils.NewSubPath("tools", ArduinoDataFolder, filepath.Join("staging", item), true)
}

// IndexPath returns the path of the specified index file.
func IndexPath(fileName string) pathutils.Path {
	return pathutils.NewSubPath(fileName, ArduinoDataFolder, fileName, false)
}

func getDefaultArduinoDataFolder() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("retrieving home dir: %s", err)
	}
	arduinoDataFolder := usr.HomeDir

	switch runtime.GOOS {
	case "linux":
		arduinoDataFolder = filepath.Join(arduinoDataFolder, ".arduino15")
		break
	case "darwin":
		arduinoDataFolder = filepath.Join(arduinoDataFolder, "Library", "arduino15")
		break
	case "windows":
		return "", fmt.Errorf("Windows temporarily unsupported")
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return arduinoDataFolder, nil
}

func getDefaultSketchbookFolder() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("retrieving home dir: %s", err)
	}
	// TODO: before doing this, check IDE's preferences.txt for different sketchbook path
	return filepath.Join(usr.HomeDir, "Arduino"), nil
}
