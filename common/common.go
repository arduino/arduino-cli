package common

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// GetDefaultArduinoFolder returns the default data folder for Arduino platform
func GetDefaultArduinoFolder() (string, error) {
	var folder string

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "linux":
		folder = filepath.Join(usr.HomeDir, ".arduino15")
	case "darwin":
		folder = filepath.Join(usr.HomeDir, "Library", "arduino15")
	default:
		return "", fmt.Errorf("Unsupported OS: %s", runtime.GOOS)
	}
	_, err = os.Stat(folder)
	if os.IsNotExist(err) {
		fmt.Print("Cannot find default arduino folder, attemping to create it ...")
		err = os.MkdirAll(folder, 0755)
		if err != nil {
			fmt.Println("ERROR")
			fmt.Println("Cannot create arduino folder")
			return "", err
		}
		fmt.Println("OK")
	}
	return folder, nil
}

// GetDefaultLibFolder get the default folder of downloaded libraries.
func GetDefaultLibFolder() (string, error) {
	baseFolder, err := GetDefaultArduinoFolder()
	if err != nil {
		return "", err
	}

	libFolder := filepath.Join(baseFolder, "libraries")
	_, err = os.Stat(libFolder)
	if os.IsNotExist(err) {
		fmt.Print("Cannot find libraries folder, attempting to create it ... ")
		err = os.MkdirAll(libFolder, 0755)
		if err != nil {
			fmt.Println("ERROR")
			fmt.Println("Cannot create libraries folder")
			return "", err
		}
		fmt.Println("OK")
	}
	return libFolder, nil
}

func Unzip(archive *zip.Reader, destination string) error {
	for _, file := range archive.File {
		path := filepath.Join(destination, file.Name)
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return fmt.Errorf("Cannot create directory during extraction. Process has been aborted")
			}
		} else {
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				return fmt.Errorf("Cannot create directory tree of file during extraction. Process has been aborted")
			}

			fileOpened, err := file.Open()
			if err != nil {
				return fmt.Errorf("Cannot open archived file, process has been aborted")
			}
			content, err := ioutil.ReadAll(fileOpened)
			if err != nil {
				return fmt.Errorf("Cannot read archived file, process has been aborted")
			}
			err = ioutil.WriteFile(path, content, 0664)
			if err != nil {
				return fmt.Errorf("Cannot copy archived file, process has been aborted, %s", err)
			}
		}
	}
	return nil
}
