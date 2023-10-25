// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

package result

import (
	"cmp"

	"github.com/arduino/arduino-cli/internal/orderedmap"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	semver "go.bug.st/relaxed-semver"
)

// NewPlatformSummary creates a new result.Platform from rpc.PlatformSummary
func NewPlatformSummary(in *rpc.PlatformSummary) *PlatformSummary {
	releases := orderedmap.NewWithConversionFunc[*semver.Version, *PlatformRelease, string]((*semver.Version).String)
	for k, v := range in.Releases {
		releases.Set(semver.MustParse(k), NewPlatformReleaseResult(v))
	}
	releases.SortKeys((*semver.Version).CompareTo)

	return &PlatformSummary{
		Id:                in.Metadata.Id,
		Maintainer:        in.Metadata.Maintainer,
		Website:           in.Metadata.Website,
		Email:             in.Metadata.Email,
		ManuallyInstalled: in.Metadata.ManuallyInstalled,
		Deprecated:        in.Metadata.Deprecated,
		Indexed:           in.Metadata.Indexed,
		Releases:          releases,
		InstalledVersion:  semver.MustParse(in.InstalledVersion),
		LatestVersion:     semver.MustParse(in.LatestVersion),
	}
}

// PlatformSummary maps a rpc.PlatformSummary
type PlatformSummary struct {
	Id                string `json:"id,omitempty"`
	Maintainer        string `json:"maintainer,omitempty"`
	Website           string `json:"website,omitempty"`
	Email             string `json:"email,omitempty"`
	ManuallyInstalled bool   `json:"manually_installed,omitempty"`
	Deprecated        bool   `json:"deprecated,omitempty"`
	Indexed           bool   `json:"indexed,omitempty"`

	Releases orderedmap.Map[*semver.Version, *PlatformRelease] `json:"releases,omitempty"`

	InstalledVersion *semver.Version `json:"installed_version,omitempty"`
	LatestVersion    *semver.Version `json:"latest_version,omitempty"`
}

// GetLatestRelease returns the latest relase of this platform or nil if none available.
func (p *PlatformSummary) GetLatestRelease() *PlatformRelease {
	return p.Releases.Get(p.LatestVersion)
}

// GetInstalledRelease returns the installed relase of this platform or nil if none available.
func (p *PlatformSummary) GetInstalledRelease() *PlatformRelease {
	return p.Releases.Get(p.InstalledVersion)
}

// NewPlatformReleaseResult creates a new result.PlatformRelease from rpc.PlatformRelease
func NewPlatformReleaseResult(in *rpc.PlatformRelease) *PlatformRelease {
	var boards []*Board
	for _, board := range in.Boards {
		boards = append(boards, &Board{
			Name: board.Name,
			Fqbn: board.Fqbn,
		})
	}
	var help *HelpResource
	if in.Help != nil {
		help = &HelpResource{
			Online: in.Help.Online,
		}
	}
	res := &PlatformRelease{
		Name:            in.Name,
		Version:         in.Version,
		Type:            in.Type,
		Installed:       in.Installed,
		Boards:          boards,
		Help:            help,
		MissingMetadata: in.MissingMetadata,
		Deprecated:      in.Deprecated,
	}
	return res
}

// PlatformRelease maps a rpc.PlatformRelease
type PlatformRelease struct {
	Name            string        `json:"name,omitempty"`
	Version         string        `json:"version,omitempty"`
	Type            []string      `json:"type,omitempty"`
	Installed       bool          `json:"installed,omitempty"`
	Boards          []*Board      `json:"boards,omitempty"`
	Help            *HelpResource `json:"help,omitempty"`
	MissingMetadata bool          `json:"missing_metadata,omitempty"`
	Deprecated      bool          `json:"deprecated,omitempty"`
}

// Board maps a rpc.Board
type Board struct {
	Name string `json:"name,omitempty"`
	Fqbn string `json:"fqbn,omitempty"`
}

// HelpResource maps a rpc.HelpResource
type HelpResource struct {
	Online string `json:"online,omitempty"`
}

type InstalledLibrary struct {
	Library *Library        `json:"library,omitempty"`
	Release *LibraryRelease `json:"release,omitempty"`
}

type LibraryLocation string

type LibraryLayout string

type Library struct {
	Name              string                         `json:"name,omitempty"`
	Author            string                         `json:"author,omitempty"`
	Maintainer        string                         `json:"maintainer,omitempty"`
	Sentence          string                         `json:"sentence,omitempty"`
	Paragraph         string                         `json:"paragraph,omitempty"`
	Website           string                         `json:"website,omitempty"`
	Category          string                         `json:"category,omitempty"`
	Architectures     []string                       `json:"architectures,omitempty"`
	Types             []string                       `json:"types,omitempty"`
	InstallDir        string                         `json:"install_dir,omitempty"`
	SourceDir         string                         `json:"source_dir,omitempty"`
	UtilityDir        string                         `json:"utility_dir,omitempty"`
	ContainerPlatform string                         `json:"container_platform,omitempty"`
	DotALinkage       bool                           `json:"dot_a_linkage,omitempty"`
	Precompiled       bool                           `json:"precompiled,omitempty"`
	LdFlags           string                         `json:"ld_flags,omitempty"`
	IsLegacy          bool                           `json:"is_legacy,omitempty"`
	Version           string                         `json:"version,omitempty"`
	License           string                         `json:"license,omitempty"`
	Properties        orderedmap.Map[string, string] `json:"properties,omitempty"`
	Location          LibraryLocation                `json:"location,omitempty"`
	Layout            LibraryLayout                  `json:"layout,omitempty"`
	Examples          []string                       `json:"examples,omitempty"`
	ProvidesIncludes  []string                       `json:"provides_includes,omitempty"`
	CompatibleWith    orderedmap.Map[string, bool]   `json:"compatible_with,omitempty"`
	InDevelopment     bool                           `json:"in_development,omitempty"`
}

type LibraryRelease struct {
	Author           string               `json:"author,omitempty"`
	Version          string               `json:"version,omitempty"`
	Maintainer       string               `json:"maintainer,omitempty"`
	Sentence         string               `json:"sentence,omitempty"`
	Paragraph        string               `json:"paragraph,omitempty"`
	Website          string               `json:"website,omitempty"`
	Category         string               `json:"category,omitempty"`
	Architectures    []string             `json:"architectures,omitempty"`
	Types            []string             `json:"types,omitempty"`
	Resources        *DownloadResource    `json:"resources,omitempty"`
	License          string               `json:"license,omitempty"`
	ProvidesIncludes []string             `json:"provides_includes,omitempty"`
	Dependencies     []*LibraryDependency `json:"dependencies,omitempty"`
}

type DownloadResource struct {
	Url             string `json:"url,omitempty"`
	ArchiveFilename string `json:"archive_filename,omitempty"`
	Checksum        string `json:"checksum,omitempty"`
	Size            int64  `json:"size,omitempty"`
	CachePath       string `json:"cache_path,omitempty"`
}

type LibraryDependency struct {
	Name              string `json:"name,omitempty"`
	VersionConstraint string `json:"version_constraint,omitempty"`
}

func NewInstalledLibraryResult(l *rpc.InstalledLibrary) *InstalledLibrary {
	libraryPropsMap := orderedmap.New[string, string]()
	for k, v := range l.GetLibrary().GetProperties() {
		libraryPropsMap.Set(k, v)
	}
	libraryPropsMap.SortStableKeys(cmp.Compare)

	libraryCompatibleWithMap := orderedmap.New[string, bool]()
	for k, v := range l.GetLibrary().GetCompatibleWith() {
		libraryCompatibleWithMap.Set(k, v)
	}
	libraryCompatibleWithMap.SortStableKeys(cmp.Compare)

	return &InstalledLibrary{
		Library: &Library{
			Name:              l.GetLibrary().GetName(),
			Author:            l.GetLibrary().GetAuthor(),
			Maintainer:        l.GetLibrary().GetMaintainer(),
			Sentence:          l.GetLibrary().GetSentence(),
			Paragraph:         l.GetLibrary().GetParagraph(),
			Website:           l.GetLibrary().GetWebsite(),
			Category:          l.GetLibrary().GetCategory(),
			Architectures:     l.GetLibrary().GetArchitectures(),
			Types:             l.GetLibrary().GetTypes(),
			InstallDir:        l.GetLibrary().GetInstallDir(),
			SourceDir:         l.GetLibrary().GetSourceDir(),
			UtilityDir:        l.GetLibrary().GetUtilityDir(),
			ContainerPlatform: l.GetLibrary().GetContainerPlatform(),
			DotALinkage:       l.GetLibrary().GetDotALinkage(),
			Precompiled:       l.GetLibrary().GetPrecompiled(),
			LdFlags:           l.GetLibrary().GetLdFlags(),
			IsLegacy:          l.GetLibrary().GetIsLegacy(),
			Version:           l.GetLibrary().GetVersion(),
			License:           l.GetLibrary().GetLicense(),
			Properties:        libraryPropsMap,
			Location:          LibraryLocation(l.GetLibrary().GetLocation().String()),
			Layout:            LibraryLayout(l.GetLibrary().GetLayout().String()),
			Examples:          l.GetLibrary().GetExamples(),
			ProvidesIncludes:  l.GetLibrary().GetProvidesIncludes(),
			CompatibleWith:    libraryCompatibleWithMap,
			InDevelopment:     l.GetLibrary().GetInDevelopment(),
		},
		Release: &LibraryRelease{
			Author:           l.GetRelease().GetAuthor(),
			Version:          l.GetRelease().GetVersion(),
			Maintainer:       l.GetRelease().GetMaintainer(),
			Sentence:         l.GetRelease().GetSentence(),
			Paragraph:        l.GetRelease().GetParagraph(),
			Website:          l.GetRelease().GetWebsite(),
			Category:         l.GetRelease().GetCategory(),
			Architectures:    l.GetRelease().GetArchitectures(),
			Types:            l.GetRelease().GetTypes(),
			Resources:        NewDownloadResource(l.GetRelease().GetResources()),
			License:          l.GetRelease().GetLicense(),
			ProvidesIncludes: l.GetRelease().GetProvidesIncludes(),
			Dependencies:     NewLibraryDependencies(l.GetRelease().GetDependencies()),
		},
	}
}

func NewDownloadResource(r *rpc.DownloadResource) *DownloadResource {
	if r == nil {
		return nil
	}
	return &DownloadResource{
		Url:             r.GetUrl(),
		ArchiveFilename: r.GetArchiveFilename(),
		Checksum:        r.GetChecksum(),
		Size:            r.GetSize(),
		CachePath:       r.GetCachePath(),
	}
}

func NewLibraryDependencies(d []*rpc.LibraryDependency) []*LibraryDependency {
	if d == nil {
		return nil
	}
	result := make([]*LibraryDependency, len(d))
	for i, v := range d {
		result[i] = NewLibraryDependency(v)
	}
	return result
}

func NewLibraryDependency(d *rpc.LibraryDependency) *LibraryDependency {
	if d == nil {
		return nil
	}
	return &LibraryDependency{
		Name:              d.GetName(),
		VersionConstraint: d.GetVersionConstraint(),
	}
}

type Port struct {
	Address       string                         `json:"address,omitempty"`
	Label         string                         `json:"label,omitempty"`
	Protocol      string                         `json:"protocol,omitempty"`
	ProtocolLabel string                         `json:"protocol_label,omitempty"`
	Properties    orderedmap.Map[string, string] `json:"properties,omitempty"`
	HardwareId    string                         `json:"hardware_id,omitempty"`
}

func NewPort(p *rpc.Port) *Port {
	propertiesMap := orderedmap.New[string, string]()
	for k, v := range p.GetProperties() {
		propertiesMap.Set(k, v)
	}
	propertiesMap.SortStableKeys(cmp.Compare)
	if p == nil {
		return nil
	}
	return &Port{
		Address:       p.GetAddress(),
		Label:         p.GetLabel(),
		Protocol:      p.GetProtocol(),
		ProtocolLabel: p.GetProtocolLabel(),
		Properties:    propertiesMap,
		HardwareId:    p.GetHardwareId(),
	}
}

type BoardDetailsResponse struct {
	Fqbn                     string                           `json:"fqbn,omitempty"`
	Name                     string                           `json:"name,omitempty"`
	Version                  string                           `json:"version,omitempty"`
	PropertiesId             string                           `json:"properties_id,omitempty"`
	Alias                    string                           `json:"alias,omitempty"`
	Official                 bool                             `json:"official,omitempty"`
	Pinout                   string                           `json:"pinout,omitempty"`
	Package                  *Package                         `json:"package,omitempty"`
	Platform                 *BoardPlatform                   `json:"platform,omitempty"`
	ToolsDependencies        []*ToolsDependency               `json:"tools_dependencies,omitempty"`
	ConfigOptions            []*ConfigOption                  `json:"config_options,omitempty"`
	Programmers              []*Programmer                    `json:"programmers,omitempty"`
	DebuggingSupported       bool                             `json:"debugging_supported,omitempty"`
	IdentificationProperties []*BoardIdentificationProperties `json:"identification_properties,omitempty"`
	BuildProperties          []string                         `json:"build_properties,omitempty"`
}

func NewBoardDetailsResponse(b *rpc.BoardDetailsResponse) *BoardDetailsResponse {
	if b == nil {
		return nil
	}
	return &BoardDetailsResponse{
		Fqbn:                     b.GetFqbn(),
		Name:                     b.GetName(),
		Version:                  b.GetVersion(),
		PropertiesId:             b.GetPropertiesId(),
		Alias:                    b.GetAlias(),
		Official:                 b.GetOfficial(),
		Pinout:                   b.GetPinout(),
		Package:                  NewPackage(b.GetPackage()),
		Platform:                 NewBoardPlatform(b.GetPlatform()),
		ToolsDependencies:        NewToolsDependencies(b.GetToolsDependencies()),
		ConfigOptions:            NewConfigOptions(b.GetConfigOptions()),
		Programmers:              NewProgrammers(b.GetProgrammers()),
		DebuggingSupported:       b.GetDebuggingSupported(),
		IdentificationProperties: NewBoardIdentificationProperties(b.GetIdentificationProperties()),
		BuildProperties:          b.GetBuildProperties(),
	}
}

type Package struct {
	Maintainer string `json:"maintainer,omitempty"`
	Url        string `json:"url,omitempty"`
	WebsiteUrl string `json:"website_url,omitempty"`
	Email      string `json:"email,omitempty"`
	Name       string `json:"name,omitempty"`
	Help       *Help  `json:"help,omitempty"`
}

func NewPackage(p *rpc.Package) *Package {
	if p == nil {
		return nil
	}
	return &Package{
		Maintainer: p.GetMaintainer(),
		Url:        p.GetUrl(),
		WebsiteUrl: p.GetWebsiteUrl(),
		Email:      p.GetEmail(),
		Name:       p.GetName(),
		Help:       NewHelp(p.GetHelp()),
	}
}

type Help struct {
	Online string `json:"online,omitempty"`
}

func NewHelp(h *rpc.Help) *Help {
	if h == nil {
		return nil
	}
	return &Help{Online: h.GetOnline()}
}

type BoardPlatform struct {
	Architecture    string `json:"architecture,omitempty"`
	Category        string `json:"category,omitempty"`
	Url             string `json:"url,omitempty"`
	ArchiveFilename string `json:"archive_filename,omitempty"`
	Checksum        string `json:"checksum,omitempty"`
	Size            int64  `json:"size,omitempty"`
	Name            string `json:"name,omitempty"`
}

func NewBoardPlatform(p *rpc.BoardPlatform) *BoardPlatform {
	if p == nil {
		return nil
	}
	return &BoardPlatform{
		Architecture:    p.GetArchitecture(),
		Category:        p.GetCategory(),
		Url:             p.GetUrl(),
		ArchiveFilename: p.GetArchiveFilename(),
		Checksum:        p.GetChecksum(),
		Size:            p.GetSize(),
		Name:            p.GetName(),
	}
}

type ToolsDependency struct {
	Packager string    `json:"packager,omitempty"`
	Name     string    `json:"name,omitempty"`
	Version  string    `json:"version,omitempty"`
	Systems  []*System `json:"systems,omitempty"`
}

func NewToolsDependencies(p []*rpc.ToolsDependencies) []*ToolsDependency {
	if p == nil {
		return nil
	}
	res := make([]*ToolsDependency, len(p))
	for i, v := range p {
		res[i] = NewToolsDependency(v)
	}
	return res
}

func NewToolsDependency(p *rpc.ToolsDependencies) *ToolsDependency {
	if p == nil {
		return nil
	}
	return &ToolsDependency{
		Packager: p.GetPackager(),
		Name:     p.GetName(),
		Version:  p.GetVersion(),
		Systems:  NewSystems(p.GetSystems()),
	}
}

type System struct {
	Checksum        string `json:"checksum,omitempty"`
	Host            string `json:"host,omitempty"`
	ArchiveFilename string `json:"archive_filename,omitempty"`
	Url             string `json:"url,omitempty"`
	Size            int64  `json:"size,omitempty"`
}

func NewSystems(p []*rpc.Systems) []*System {
	if p == nil {
		return nil
	}
	res := make([]*System, len(p))
	for i, v := range p {
		res[i] = NewSystem(v)
	}
	return res
}

func NewSystem(s *rpc.Systems) *System {
	if s == nil {
		return nil
	}
	return &System{
		Checksum:        s.GetChecksum(),
		Host:            s.GetHost(),
		ArchiveFilename: s.GetArchiveFilename(),
		Url:             s.GetUrl(),
		Size:            s.GetSize(),
	}
}

type ConfigOption struct {
	Option      string         `json:"option,omitempty"`
	OptionLabel string         `json:"option_label,omitempty"`
	Values      []*ConfigValue `json:"values,omitempty"`
}

func NewConfigOptions(c []*rpc.ConfigOption) []*ConfigOption {
	if c == nil {
		return nil
	}
	res := make([]*ConfigOption, len(c))
	for i, v := range c {
		res[i] = NewConfigOption(v)
	}
	return res
}

func NewConfigOption(o *rpc.ConfigOption) *ConfigOption {
	if o == nil {
		return nil
	}
	return &ConfigOption{
		Option:      o.GetOption(),
		OptionLabel: o.GetOptionLabel(),
		Values:      NewConfigValues(o.GetValues()),
	}
}

type ConfigValue struct {
	Value      string `json:"value,omitempty"`
	ValueLabel string `json:"value_label,omitempty"`
	Selected   bool   `json:"selected,omitempty"`
}

func NewConfigValues(c []*rpc.ConfigValue) []*ConfigValue {
	if c == nil {
		return nil
	}
	res := make([]*ConfigValue, len(c))
	for i, v := range c {
		res[i] = NewConfigValue(v)
	}
	return res
}

func NewConfigValue(c *rpc.ConfigValue) *ConfigValue {
	if c == nil {
		return nil
	}
	return &ConfigValue{
		Value:      c.GetValue(),
		ValueLabel: c.GetValueLabel(),
		Selected:   c.GetSelected(),
	}
}

type Programmer struct {
	Platform string `json:"platform,omitempty"`
	Id       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
}

func NewProgrammers(c []*rpc.Programmer) []*Programmer {
	if c == nil {
		return nil
	}
	res := make([]*Programmer, len(c))
	for i, v := range c {
		res[i] = NewProgrammer(v)
	}
	return res
}

func NewProgrammer(c *rpc.Programmer) *Programmer {
	if c == nil {
		return nil
	}
	return &Programmer{
		Platform: c.GetPlatform(),
		Id:       c.GetId(),
		Name:     c.GetName(),
	}
}

type BoardIdentificationProperties struct {
	Properties orderedmap.Map[string, string] `json:"properties,omitempty"`
}

func NewBoardIdentificationProperties(p []*rpc.BoardIdentificationProperties) []*BoardIdentificationProperties {
	if p == nil {
		return nil
	}
	res := make([]*BoardIdentificationProperties, len(p))
	for i, v := range p {
		res[i] = NewBoardIndentificationProperty(v)
	}
	return res
}

func NewBoardIndentificationProperty(p *rpc.BoardIdentificationProperties) *BoardIdentificationProperties {
	if p == nil {
		return nil
	}
	propertiesMap := orderedmap.New[string, string]()
	for k, v := range p.GetProperties() {
		propertiesMap.Set(k, v)
	}
	propertiesMap.SortStableKeys(cmp.Compare)

	return &BoardIdentificationProperties{Properties: propertiesMap}
}
