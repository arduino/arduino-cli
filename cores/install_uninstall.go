package cores

import (
	"archive/zip"
	"errors"
	"io/ioutil"
	"os"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/codeclysm/extract"
)

// Install installs a specific release of a core.
func Install(packager, name string, release releases.Release) error {
	if release == nil {
		return errors.New("Not existing version of the core")
	}

	coresFolder, err := common.GetDefaultCoresFolder(packager)
	if err != nil {
		return err
	}

	cacheFilePath, err := releases.ArchivePath(release)
	if err != nil {
		return err
	}

	tempFolder, err := ioutil.TempDir("cores", name)
	if err != nil {
		return err
	}

	file, err := os.Open(cacheFilePath)
	if err != nil {
		return err
	}

	extract.Archive(file, tempFolder, nil)

	purgeTempDir()
	moveTempDir()

	return nil
}

// Install installs a library.
func Install(name string, release releases.Release) error {
	if release == nil {
		return errors.New("Not existing version of the library")
	}

	/*
		installedRelease, err := library.InstalledRelease()
		if err != nil {
			return err
		}
		if installedRelease != nil {
			//if installedRelease.Version != library.Latest().Version {
			err := removeRelease(library.Name, installedRelease)
			if err != nil {
				return err
			}
			//} else {
			//	return nil // Already installed and latest version.
			//}
		}
	*/
	libFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		return err
	}

	cacheFilePath, err := release.ArchivePath()
	if err != nil {
		return err
	}

	zipArchive, err := zip.OpenReader(cacheFilePath)
	if err != nil {
		return err
	}
	defer zipArchive.Close()

	err = common.Unzip(zipArchive, libFolder)
	if err != nil {
		return err
	}

	return nil
}
