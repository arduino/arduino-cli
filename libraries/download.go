package libraries

import (
	"archive/zip"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/common"
)

// DownloadAndCache downloads a library without installing it
func DownloadAndCache(library *Library) (*zip.Reader, error) {
	zipContent, err := DownloadLatest(library)
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
func DownloadLatest(library *Library) ([]byte, error) {
	return common.DownloadPackage(library.Latest.URL)
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
