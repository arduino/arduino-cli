package cores

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

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
	defer os.RemoveAll(tempFolder)

	file, err := os.Open(cacheFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	extract.Archive(file, tempFolder, nil)

	realDir := coreTempDir(tempFolder)
	if realDir == "invalid" {
		return errors.New("invalid archive structure")
	}

	destCoreDir := filepath.Join(coresFolder, name, release.VersionName())

	os.Rename(realDir, destCoreDir)

	return nil
}

// InstallTool installs a specific release of a tool.
func InstallTool(packager, name string, release releases.Release) error {
	if release == nil {
		return errors.New("Not existing version of the core")
	}

	toolsFolder, err := common.GetDefaultToolsFolder(packager)
	if err != nil {
		return err
	}

	cacheFilePath, err := releases.ArchivePath(release)
	if err != nil {
		return err
	}

	tempFolder, err := ioutil.TempDir("tools", name)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempFolder)

	file, err := os.Open(cacheFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	extract.Archive(file, tempFolder, nil)

	realDir := toolTempDir(tempFolder)
	if realDir == "invalid" {
		return errors.New("invalid archive structure")
	}

	destToolDir := filepath.Join(toolsFolder, name, release.VersionName())

	os.Rename(realDir, destToolDir)

	return nil
}

func coreTempDir(tempDir string) string {
	realDir := "invalid"
	filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if info.Name() == "platform.txt" {
			realDir = filepath.Dir(path)
			return errors.New("stopped, ok") //error put to stop the search of the root
		}
		return nil
	})
	return realDir
}

func toolTempDir(tempDir string) string {
	realDir := "invalid"
	filepath.Walk(tempDir, func(path string) string {
		if info.Name() == "platform.txt" {
			realDir = filepath.Dir(path)
		}
		return nil
	})
	return realDir
}
