package libraries

import (
	"archive/zip"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/common"
)

const (
	libraryIndexURL string = "http://downloads.arduino.cc/libraries/library_index.json"
)

// DownloadAndCache downloads a library without installing it
func DownloadAndCache(library *Library) (*zip.Reader, error) {
	zipContent, err := downloadLatest(library)
	if err != nil {
		return nil, err
	}

	zipArchive, err := prepareInstall(library, zipContent)
	if err != nil {
		return nil, err
	}
	return zipArchive, nil
}

// DownloadLatest downloads Latest version of a library.
func downloadLatest(library *Library) ([]byte, error) {
	return common.DownloadPackage(library.Latest.URL)
}

//DownloadLibrariesFile downloads the lib file from arduino repository.
func DownloadLibrariesFile() error {
	libFile, err := IndexPath()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", libraryIndexURL, nil)
	if err != nil {
		return err
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(libFile, content, 0666)
	if err != nil {
		return err
	}
	return nil
}

// getDownloadCacheFolder gets the folder where temp installs are stored until installation complete (libraries).
func getDownloadCacheFolder(library *Library) (string, error) {
	libFolder, err := common.GetDefaultArduinoHomeFolder()
	if err != nil {
		return "", err
	}

	stagingFolder := filepath.Join(libFolder, "staging")
	return common.GetFolder(stagingFolder, "libraries cache")
}
