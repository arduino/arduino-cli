// This file is part of arduino-cli.
//
// Copyright 2020-2022 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package sketch

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/arduino/arduino-cli/internal/arduino/utils"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	semver "go.bug.st/relaxed-semver"
	"gopkg.in/yaml.v3"
)

// projectRaw is a support struct used only to unmarshal the yaml
type projectRaw struct {
	ProfilesRaw       yaml.Node `yaml:"profiles"`
	DefaultProfile    string    `yaml:"default_profile"`
	DefaultFqbn       string    `yaml:"default_fqbn"`
	DefaultPort       string    `yaml:"default_port,omitempty"`
	DefaultProtocol   string    `yaml:"default_protocol,omitempty"`
	DefaultProgrammer string    `yaml:"default_programmer,omitempty"`
}

// Project represents the sketch project file
type Project struct {
	Profiles          []*Profile
	DefaultProfile    string
	DefaultFqbn       string
	DefaultPort       string
	DefaultProtocol   string
	DefaultProgrammer string
}

// AsYaml outputs the sketch project file as YAML
func (p *Project) AsYaml() string {
	res := "profiles:\n"

	for _, profile := range p.Profiles {
		res += fmt.Sprintf("  %s:\n", profile.Name)
		res += profile.AsYaml()
		res += "\n"
	}
	if p.DefaultProfile != "" {
		res += fmt.Sprintf("default_profile: %s\n", p.DefaultProfile)
	}
	if p.DefaultFqbn != "" {
		res += fmt.Sprintf("default_fqbn: %s\n", p.DefaultFqbn)
	}
	if p.DefaultPort != "" {
		res += fmt.Sprintf("default_port: %s\n", p.DefaultPort)
	}
	if p.DefaultProtocol != "" {
		res += fmt.Sprintf("default_protocol: %s\n", p.DefaultProtocol)
	}
	if p.DefaultProgrammer != "" {
		res += fmt.Sprintf("default_programmer: %s\n", p.DefaultProgrammer)
	}
	return res
}

func (p *projectRaw) getProfiles() ([]*Profile, error) {
	profiles := []*Profile{}
	for i, node := range p.ProfilesRaw.Content {
		if node.Tag != "!!str" {
			continue // Node is a map, so it is read out at key.
		}

		var profile Profile
		profile.Name = node.Value
		if err := p.ProfilesRaw.Content[i+1].Decode(&profile); err != nil {
			return nil, err
		}
		profiles = append(profiles, &profile)
	}
	return profiles, nil
}

// UnmarshalYAML decodes a Profiles section from YAML source.
// Profile is a sketch profile, it contains a reference to all the resources
// needed to build and upload a sketch
type Profile struct {
	Name       string
	Notes      string                   `yaml:"notes"`
	FQBN       string                   `yaml:"fqbn"`
	Programmer string                   `yaml:"programmer"`
	Platforms  ProfileRequiredPlatforms `yaml:"platforms"`
	Libraries  ProfileRequiredLibraries `yaml:"libraries"`
}

// ToRpc converts this Profile to an rpc.SketchProfile
func (p *Profile) ToRpc() *rpc.SketchProfile {
	return &rpc.SketchProfile{
		Name:       p.Name,
		Fqbn:       p.FQBN,
		Programmer: p.Programmer,
	}
}

// AsYaml outputs the profile as Yaml
func (p *Profile) AsYaml() string {
	res := ""
	if p.Notes != "" {
		res += fmt.Sprintf("    notes: %s\n", p.Notes)
	}
	res += fmt.Sprintf("    fqbn: %s\n", p.FQBN)
	if p.Programmer != "" {
		res += fmt.Sprintf("    programmer: %s\n", p.Programmer)
	}
	res += p.Platforms.AsYaml()
	res += p.Libraries.AsYaml()
	return res
}

// ProfileRequiredPlatforms is a list of ProfilePlatformReference (platforms
// required to build the sketch using this profile)
type ProfileRequiredPlatforms []*ProfilePlatformReference

// AsYaml outputs the required platforms as Yaml
func (p *ProfileRequiredPlatforms) AsYaml() string {
	res := "    platforms:\n"
	for _, platform := range *p {
		res += platform.AsYaml()
	}
	return res
}

// ProfileRequiredLibraries is a list of ProfileLibraryReference (libraries
// required to build the sketch using this profile)
type ProfileRequiredLibraries []*ProfileLibraryReference

// AsYaml outputs the required libraries as Yaml
func (p *ProfileRequiredLibraries) AsYaml() string {
	res := "    libraries:\n"
	for _, lib := range *p {
		res += lib.AsYaml()
	}
	return res
}

// ProfilePlatformReference is a reference to a platform
type ProfilePlatformReference struct {
	Packager         string
	Architecture     string
	Version          *semver.Version
	PlatformIndexURL *url.URL
}

// InternalUniqueIdentifier returns the unique identifier for this object
func (p *ProfilePlatformReference) InternalUniqueIdentifier() string {
	id := p.String()
	h := sha256.Sum256([]byte(id))
	res := fmt.Sprintf("%s:%s@%s_%s", p.Packager, p.Architecture, p.Version, hex.EncodeToString(h[:])[:16])
	return utils.SanitizeName(res)
}

func (p *ProfilePlatformReference) String() string {
	res := fmt.Sprintf("%s:%s@%s", p.Packager, p.Architecture, p.Version)
	if p.PlatformIndexURL != nil {
		res += fmt.Sprintf(" (%s)", p.PlatformIndexURL)
	}
	return res
}

// AsYaml outputs the platform reference as Yaml
func (p *ProfilePlatformReference) AsYaml() string {
	res := fmt.Sprintf("      - platform: %s:%s (%s)\n", p.Packager, p.Architecture, p.Version)
	if p.PlatformIndexURL != nil {
		res += fmt.Sprintf("        platform_index_url: %s\n", p.PlatformIndexURL)
	}
	return res
}

func parseNameAndVersion(in string) (string, string, bool) {
	re := regexp.MustCompile(`^([a-zA-Z0-9.\-_ :]+) \((.+)\)$`)
	split := re.FindAllStringSubmatch(in, -1)
	if len(split) != 1 || len(split[0]) != 3 {
		return "", "", false
	}
	return split[0][1], split[0][2], true
}

// UnmarshalYAML decodes a ProfilePlatformReference from YAML source.
func (p *ProfilePlatformReference) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var data map[string]string
	if err := unmarshal(&data); err != nil {
		return err
	}
	if platformID, ok := data["platform"]; !ok {
		return errors.New(i18n.Tr("missing '%s' directive", "platform"))
	} else if platformID, platformVersion, ok := parseNameAndVersion(platformID); !ok {
		return errors.New(i18n.Tr("invalid '%s' directive", "platform"))
	} else if c, err := semver.Parse(platformVersion); err != nil {
		return fmt.Errorf("%s: %w", i18n.Tr("error parsing version constraints"), err)
	} else if split := strings.SplitN(platformID, ":", 2); len(split) != 2 {
		return fmt.Errorf("%s: %s", i18n.Tr("invalid platform identifier"), platformID)
	} else {
		p.Packager = split[0]
		p.Architecture = split[1]
		p.Version = c
	}

	if rawIndexURL, ok := data["platform_index_url"]; ok {
		indexURL, err := url.Parse(rawIndexURL)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.Tr("invalid platform index URL:"), err)
		}
		p.PlatformIndexURL = indexURL
	}
	return nil
}

// ProfileLibraryReference is a reference to a library
type ProfileLibraryReference struct {
	Library string
	Version *semver.Version
}

// UnmarshalYAML decodes a ProfileLibraryReference from YAML source.
func (l *ProfileLibraryReference) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var data string
	if err := unmarshal(&data); err != nil {
		return err
	}
	if libName, libVersion, ok := parseNameAndVersion(data); !ok {
		return fmt.Errorf("%s %s", i18n.Tr("invalid library directive:"), data)
	} else if v, err := semver.Parse(libVersion); err != nil {
		return fmt.Errorf("%s %w", i18n.Tr("invalid version:"), err)
	} else {
		l.Library = libName
		l.Version = v
	}
	return nil
}

// AsYaml outputs the required library as Yaml
func (l *ProfileLibraryReference) AsYaml() string {
	res := fmt.Sprintf("      - %s (%s)\n", l.Library, l.Version)
	return res
}

func (l *ProfileLibraryReference) String() string {
	return fmt.Sprintf("%s@%s", l.Library, l.Version)
}

// InternalUniqueIdentifier returns the unique identifier for this object
func (l *ProfileLibraryReference) InternalUniqueIdentifier() string {
	id := l.String()
	h := sha256.Sum256([]byte(id))
	res := fmt.Sprintf("%s_%s", id, hex.EncodeToString(h[:])[:16])
	return utils.SanitizeName(res)
}

// LoadProjectFile reads a sketch project file
func LoadProjectFile(file *paths.Path) (*Project, error) {
	data, err := file.ReadFile()
	if err != nil {
		return nil, err
	}
	raw := &projectRaw{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	profiles, err := raw.getProfiles()
	if err != nil {
		return nil, err
	}
	return &Project{
		Profiles:          profiles,
		DefaultProfile:    raw.DefaultProfile,
		DefaultFqbn:       raw.DefaultFqbn,
		DefaultPort:       raw.DefaultPort,
		DefaultProtocol:   raw.DefaultProtocol,
		DefaultProgrammer: raw.DefaultProgrammer,
	}, nil
}
