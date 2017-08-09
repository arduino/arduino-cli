package cores

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/blang/semver"
)

// Tool represents a single Tool, part of a Package.
type Tool struct {
	Name     string                  `json:"name,required"` // The Name of the Tool.
	Releases map[string]*ToolRelease `json:"releases"`      //Maps Version to Release.
}

// ToolRelease represents a single release of a tool
type ToolRelease struct {
	Version  string     `json:"version,required"` // The version number of this Release.
	Flavours []*Flavour `json:"systems"`          // Maps OS to Flavour
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

// Raspberry PI, BBB or other ARM based host

// PI: "arm-linux-gnueabihf"
// Arch-linux on PI2: "armv7l-unknown-linux-gnueabihf"
// Raspbian on PI2: "arm-linux-gnueabihf"
// Ubuntu Mate on PI2: "arm-linux-gnueabihf"
// Debian 7.9 on BBB: "arm-linux-gnueabihf"
// Raspbian on PI Zero: "arm-linux-gnueabihf"
var (
	regexpArmLinux = regexp.MustCompile("arm.*-linux-gnueabihf")
	regexpAmd64    = regexp.MustCompile("x86_64-.*linux-gnu")
	regexpi386     = regexp.MustCompile("i[3456]86-.*linux-gnu")
	regexpWindows  = regexp.MustCompile("i[3456]86-.*(mingw32|cygwin)")
	regexpMac64Bit = regexp.MustCompile("(i[3456]86|x86_64)-apple-darwin.*")
	regexpmac32Bit = regexp.MustCompile("i[3456]86-apple-darwin.*")
	regexpArmBSD   = regexp.MustCompile("arm.*-freebsd[0-9]*")
)

func (f Flavour) isCompatibleWithCurrentMachine() bool {
	osName := runtime.GOOS
	osArch := runtime.GOARCH

	if f.OS == "all" {
		return true
	}

	if strings.Contains(osName, "linux") {
		if osArch == "arm" {
			return regexpArmLinux.MatchString(f.OS)
		} else if strings.Contains(osArch, "amd64") {
			return regexpAmd64.MatchString(f.OS)
		} else {
			return regexpi386.MatchString(f.OS)
		}
	} else if strings.Contains(osName, "windows") {
		return regexpWindows.MatchString(f.OS)
	} else if strings.Contains(osName, "darwin") {
		if strings.Contains(osArch, "x84_64") {
			return regexpMac64Bit.MatchString(f.OS)
		}
		return regexpmac32Bit.MatchString(f.OS)
	} else if strings.Contains(osName, "freebsd") {
		if osArch == "arm" {
			return regexpArmBSD.MatchString(f.OS)
		}
		genericFreeBSDexp := regexp.MustCompile(fmt.Sprintf("%s-freebsd[0-9]*", osArch))
		return genericFreeBSDexp.MatchString(f.OS)
	}
	return false
}

func (tr ToolRelease) getCompatibleFlavour() *Flavour {
	for _, flavour := range tr.Flavours {
		if flavour.isCompatibleWithCurrentMachine() {
			return flavour
		}
	}
	return nil
}

// Release interface implementation

// ArchiveName returns the archive file name (not the path).
func (tr ToolRelease) ArchiveName() string {
	f := tr.getCompatibleFlavour()
	if f == nil {
		return "INVALID"
	}
	return f.ArchiveFileName
}

// ArchiveURL returns the archive URL.
func (tr ToolRelease) ArchiveURL() string {
	f := tr.getCompatibleFlavour()
	if f == nil {
		return "INVALID"
	}
	return f.URL
}

// ExpectedChecksum returns the expected checksum for this release.
func (tr ToolRelease) ExpectedChecksum() string {
	f := tr.getCompatibleFlavour()
	if f == nil {
		return "INVALID"
	}
	return f.Checksum
}

// ArchiveSize returns the archive size.
func (tr ToolRelease) ArchiveSize() int64 {
	f := tr.getCompatibleFlavour()
	if f == nil {
		return -1
	}
	return f.Size
}
