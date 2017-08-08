package cores

import (
	"fmt"
	"runtime"

	"github.com/blang/semver"
)

// Tool represents a single Tool, part of a Package.
type Tool struct {
	Name     string                  `json:"name,required"` // The Name of the Tool.
	Releases map[string]*ToolRelease `json:"releases"`      //Maps Version to Release.
}

// ToolRelease represents a single release of a tool
type ToolRelease struct {
	Version  string              `json:"version,required"` // The version number of this Release.
	Flavours map[string]*Flavour `json:"systems"`          // Maps OS to Flavour
}

// Flavour represents a flavour of a Tool version.
type Flavour struct {
	OS              string `json:"os,required"`              // The OS Supported by this flavour.
	URL             string `json:"url,required"`             // The URL where to download this flavour.
	ArchiveFileName string `json:"archiveFileName,required"` // The name of the archive to download.
	Size            int64  `json:"size,required"`            // The size of the archive.
	Checksum        string `json:"checksum,required"`        // The checksum of the archive. Made like ALGO:checksum.
}

// GetVersion returns the specified release corresponding the provided version,
// or nil if not found.
func (tool Tool) GetVersion(version string) *ToolRelease {
	return tool.Releases[version]
}

// GetFlavor returns the flavor of the specified OS.
func (tr ToolRelease) GetFlavor(OS string) *Flavour {
	return tr.Flavours[OS]
}

// Versions returns all the version numbers in this Core Package.
func (tool Tool) Versions() semver.Versions {
	releases := tool.Releases
	versions := make(semver.Versions, 0, len(releases))
	for _, release := range releases {
		temp, err := semver.Make(release.Version)
		if err == nil {
			versions = append(versions, temp)
		}
	}

	return versions
}

// Latest obtains latest version of a core package.
func (tool Tool) Latest() *ToolRelease {
	return tool.GetVersion(tool.latestVersion())
}

// latestVersion obtains latest version number.
//
// It uses lexicographics to compare version strings.
func (tool Tool) latestVersion() string {
	versions := tool.Versions()
	if len(versions) > 0 {
		max := versions[0]
		for i := 1; i < len(versions); i++ {
			if versions[i].GT(max) {
				max = versions[i]
			}
		}
		return fmt.Sprint(max)
	}
	return ""
}

func (tool Tool) String() string {
	res := fmt.Sprintln("Name        :", tool.Name)
	if tool.Releases != nil && len(tool.Releases) > 0 {
		res += "Releases:\n"
		for _, release := range tool.Releases {
			res += fmt.Sprintln(release)
		}
	}
	return res
}

func (tr ToolRelease) String() string {
	res := fmt.Sprintln("  Version :", tr.Version)
	for _, f := range tr.Flavours {
		res += fmt.Sprintln(f)
	}
	return res
}

func (f Flavour) String() string {
	return fmt.Sprintln("    OS :", f.OS) +
		fmt.Sprintln("    URL:", f.URL) +
		fmt.Sprintln("    ArchiveFileName:", f.ArchiveFileName) +
		fmt.Sprintln("    Size:", f.Size) +
		fmt.Sprintln("    Checksum:", f.Checksum)
}

// Release interface implementation

// ArchiveName returns the archive file name (not the path).
func (tr ToolRelease) ArchiveName() string {
	return tr.Flavours[runtime.GOOS].ArchiveFileName
}

// ArchiveURL returns the archive URL.
func (tr ToolRelease) ArchiveURL() string {
	return tr.Flavours[runtime.GOOS].URL
}

// ExpectedChecksum returns the expected checksum for this release.
func (tr ToolRelease) ExpectedChecksum() string {
	return tr.Flavours[runtime.GOOS].Checksum
}

// ArchiveSize returns the archive size.
func (tr ToolRelease) ArchiveSize() int64 {
	return tr.Flavours[runtime.GOOS].Size
}
