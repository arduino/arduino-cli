package packagemanager_test

import (
	"encoding/json"
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores/packageindex"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

func TestInstallPlatform(t *testing.T) {
	dataDir := paths.New("testdata", "data_dir_1")
	packageDir := paths.TempDir().Join("test", "packages")
	downloadDir := paths.TempDir().Join("test", "staging")
	tmpDir := paths.TempDir().Join("test", "tmp")
	packageDir.MkdirAll()
	downloadDir.MkdirAll()
	tmpDir.MkdirAll()
	defer paths.TempDir().Join("test").RemoveAll()

	pm := packagemanager.NewPackageManager(dataDir, packageDir, downloadDir, tmpDir)
	pm.LoadPackageIndexFromFile(dataDir.Join("package_index.json"))

	platformRelease, tools, err := pm.FindPlatformReleaseDependencies(&packagemanager.PlatformReference{
		Package:              "arduino",
		PlatformArchitecture: "avr",
		PlatformVersion:      semver.MustParse("1.6.23"),
	})
	require.NotNil(t, platformRelease)
	require.NotNil(t, tools)
	require.Nil(t, err)

	downloaderConfig, err := commands.GetDownloaderConfig()
	require.NotNil(t, downloaderConfig)
	require.Nil(t, err)
	downloader, err := pm.DownloadPlatformRelease(platformRelease, downloaderConfig)
	require.NotNil(t, downloader)
	require.Nil(t, err)
	err = commands.Download(downloader, platformRelease.String(), output.NewNullDownloadProgressCB())
	require.Nil(t, err)

	err = pm.InstallPlatform(platformRelease)
	require.Nil(t, err)

	destDir := packageDir.Join("arduino", "hardware", "avr", "1.6.23")
	require.True(t, destDir.IsDir())

	installedJSON := destDir.Join("installed.json")
	require.True(t, installedJSON.Exist())

	bt, err := installedJSON.ReadFile()
	require.Nil(t, err)

	index := &packageindex.Index{}
	err = json.Unmarshal(bt, index)
	require.Nil(t, err)

	expectedJSON := paths.New("testdata", "installed.json")
	expectedBt, err := expectedJSON.ReadFile()
	require.Nil(t, err)

	expectedIndex := &packageindex.Index{}
	err = json.Unmarshal(expectedBt, expectedIndex)
	require.Nil(t, err)

	require.Equal(t, expectedIndex.IsTrusted, index.IsTrusted)
	require.Equal(t, len(expectedIndex.Packages), len(index.Packages))

	for i := range expectedIndex.Packages {
		expectedPackage := expectedIndex.Packages[i]
		indexPackage := index.Packages[i]
		require.Equal(t, expectedPackage.Name, indexPackage.Name)
		require.Equal(t, expectedPackage.Maintainer, indexPackage.Maintainer)
		require.Equal(t, expectedPackage.WebsiteURL, indexPackage.WebsiteURL)
		require.Equal(t, expectedPackage.Email, indexPackage.Email)
		require.Equal(t, expectedPackage.Help.Online, indexPackage.Help.Online)
		require.Equal(t, len(expectedPackage.Tools), len(indexPackage.Tools))
		require.ElementsMatch(t, expectedPackage.Tools, indexPackage.Tools)

		require.Equal(t, len(expectedPackage.Platforms), len(indexPackage.Platforms))
		for n := range expectedPackage.Platforms {
			expectedPlatform := expectedPackage.Platforms[n]
			indexPlatform := indexPackage.Platforms[n]
			require.Equal(t, expectedPlatform.Name, indexPlatform.Name)
			require.Equal(t, expectedPlatform.Architecture, indexPlatform.Architecture)
			require.Equal(t, expectedPlatform.Version.String(), indexPlatform.Version.String())
			require.Equal(t, expectedPlatform.Category, indexPlatform.Category)
			require.Equal(t, expectedPlatform.Help.Online, indexPlatform.Help.Online)
			require.Equal(t, expectedPlatform.URL, indexPlatform.URL)
			require.Equal(t, expectedPlatform.ArchiveFileName, indexPlatform.ArchiveFileName)
			require.Equal(t, expectedPlatform.Checksum, indexPlatform.Checksum)
			require.Equal(t, expectedPlatform.Size, indexPlatform.Size)
			require.ElementsMatch(t, expectedPlatform.Boards, indexPlatform.Boards)
			require.ElementsMatch(t, expectedPlatform.ToolDependencies, indexPlatform.ToolDependencies)
		}
	}
}
