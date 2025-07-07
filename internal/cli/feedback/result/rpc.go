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
	"fmt"
	"slices"

	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/internal/orderedmap"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"go.bug.st/f"
	semver "go.bug.st/relaxed-semver"
)

// NewPlatformSummary creates a new result.PlatformSummary from rpc.PlatformSummary
func NewPlatformSummary(in *rpc.PlatformSummary) *PlatformSummary {
	if in == nil {
		return nil
	}

	releases := orderedmap.NewWithConversionFunc[*semver.Version, *PlatformRelease, string]((*semver.Version).String)
	for k, v := range in.GetReleases() {
		releases.Set(semver.MustParse(k), NewPlatformRelease(v))
	}
	releases.SortKeys((*semver.Version).CompareTo)

	return &PlatformSummary{
		Id:                in.GetMetadata().GetId(),
		Maintainer:        in.GetMetadata().GetMaintainer(),
		Website:           in.GetMetadata().GetWebsite(),
		Email:             in.GetMetadata().GetEmail(),
		ManuallyInstalled: in.GetMetadata().GetManuallyInstalled(),
		Deprecated:        in.GetMetadata().GetDeprecated(),
		Indexed:           in.GetMetadata().GetIndexed(),
		Releases:          releases,
		InstalledVersion:  semver.MustParse(in.GetInstalledVersion()),
		LatestVersion:     semver.MustParse(in.GetLatestVersion()),
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

// GetPlatformName compute the name of the platform based on the installed/available releases.
func (p *PlatformSummary) GetPlatformName() string {
	var name string
	if installed := p.GetInstalledRelease(); installed != nil {
		name = installed.FormatName()
	}
	if name == "" {
		if latest := p.GetLatestRelease(); latest != nil {
			name = latest.FormatName()
		} else {
			keys := p.Releases.Keys()
			name = p.Releases.Get(keys[len(keys)-1]).FormatName()
		}
	}
	return name
}

// NewPlatformRelease creates a new result.PlatformRelease from rpc.PlatformRelease
func NewPlatformRelease(in *rpc.PlatformRelease) *PlatformRelease {
	if in == nil {
		return nil
	}
	var boards []*Board
	for _, board := range in.GetBoards() {
		boards = append(boards, &Board{
			Name: board.GetName(),
			Fqbn: board.GetFqbn(),
		})
	}
	var help *HelpResource
	if in.GetHelp() != nil {
		help = &HelpResource{
			Online: in.GetHelp().GetOnline(),
		}
	}
	res := &PlatformRelease{
		Name:            in.GetName(),
		Version:         in.GetVersion(),
		Types:           in.GetTypes(),
		Installed:       in.GetInstalled(),
		Boards:          boards,
		Help:            help,
		MissingMetadata: in.GetMissingMetadata(),
		Deprecated:      in.GetDeprecated(),
		Compatible:      in.GetCompatible(),
	}
	return res
}

// PlatformRelease maps a rpc.PlatformRelease
type PlatformRelease struct {
	Name            string        `json:"name,omitempty"`
	Version         string        `json:"version,omitempty"`
	Types           []string      `json:"types,omitempty"`
	Installed       bool          `json:"installed,omitempty"`
	Boards          []*Board      `json:"boards,omitempty"`
	Help            *HelpResource `json:"help,omitempty"`
	MissingMetadata bool          `json:"missing_metadata,omitempty"`
	Deprecated      bool          `json:"deprecated,omitempty"`
	Compatible      bool          `json:"compatible"`
}

func (p *PlatformRelease) FormatName() string {
	if p.Deprecated {
		return fmt.Sprintf("[%s] %s", i18n.Tr("DEPRECATED"), p.Name)
	}
	return p.Name
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

const (
	LibraryLocationUser                      LibraryLocation = "user"
	LibraryLocationIDEBuiltin                LibraryLocation = "ide"
	LibraryLocationPlatformBuiltin           LibraryLocation = "platform"
	LibraryLocationReferencedPlatformBuiltin LibraryLocation = "ref-platform"
	LibraryLocationUnmanged                  LibraryLocation = "unmanaged"
	LibraryLocationProfile                   LibraryLocation = "profile"
)

func NewLibraryLocation(r rpc.LibraryLocation) LibraryLocation {
	switch r {
	case rpc.LibraryLocation_LIBRARY_LOCATION_BUILTIN:
		return LibraryLocationIDEBuiltin
	case rpc.LibraryLocation_LIBRARY_LOCATION_PLATFORM_BUILTIN:
		return LibraryLocationPlatformBuiltin
	case rpc.LibraryLocation_LIBRARY_LOCATION_REFERENCED_PLATFORM_BUILTIN:
		return LibraryLocationReferencedPlatformBuiltin
	case rpc.LibraryLocation_LIBRARY_LOCATION_USER:
		return LibraryLocationUser
	case rpc.LibraryLocation_LIBRARY_LOCATION_UNMANAGED:
		return LibraryLocationUnmanged
	case rpc.LibraryLocation_LIBRARY_LOCATION_PROFILE:
		return LibraryLocationProfile
	}
	return LibraryLocationIDEBuiltin
}

type LibraryLayout string

const (
	LibraryLayoutFlat      LibraryLayout = "flat"
	LibraryLayoutRecursive LibraryLayout = "recursive"
)

func NewLibraryLayout(r rpc.LibraryLayout) LibraryLayout {
	switch r {
	case rpc.LibraryLayout_LIBRARY_LAYOUT_FLAT:
		return LibraryLayoutFlat
	case rpc.LibraryLayout_LIBRARY_LAYOUT_RECURSIVE:
		return LibraryLayoutRecursive
	}
	return LibraryLayoutFlat
}

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

func NewLibrary(l *rpc.Library) *Library {
	if l == nil {
		return nil
	}
	libraryPropsMap := orderedmap.New[string, string]()
	for k, v := range l.GetProperties() {
		libraryPropsMap.Set(k, v)
	}
	libraryPropsMap.SortStableKeys(cmp.Compare)

	libraryCompatibleWithMap := orderedmap.New[string, bool]()
	for k, v := range l.GetCompatibleWith() {
		libraryCompatibleWithMap.Set(k, v)
	}
	libraryCompatibleWithMap.SortStableKeys(cmp.Compare)

	return &Library{
		Name:              l.GetName(),
		Author:            l.GetAuthor(),
		Maintainer:        l.GetMaintainer(),
		Sentence:          l.GetSentence(),
		Paragraph:         l.GetParagraph(),
		Website:           l.GetWebsite(),
		Category:          l.GetCategory(),
		Architectures:     l.GetArchitectures(),
		Types:             l.GetTypes(),
		InstallDir:        l.GetInstallDir(),
		SourceDir:         l.GetSourceDir(),
		UtilityDir:        l.GetUtilityDir(),
		ContainerPlatform: l.GetContainerPlatform(),
		DotALinkage:       l.GetDotALinkage(),
		Precompiled:       l.GetPrecompiled(),
		LdFlags:           l.GetLdFlags(),
		IsLegacy:          l.GetIsLegacy(),
		Version:           l.GetVersion(),
		License:           l.GetLicense(),
		Properties:        libraryPropsMap,
		Location:          NewLibraryLocation(l.GetLocation()),
		Layout:            NewLibraryLayout(l.GetLayout()),
		Examples:          l.GetExamples(),
		ProvidesIncludes:  l.GetProvidesIncludes(),
		CompatibleWith:    libraryCompatibleWithMap,
		InDevelopment:     l.GetInDevelopment(),
	}
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

func NewLibraryRelease(l *rpc.LibraryRelease) *LibraryRelease {
	if l == nil {
		return nil
	}
	return &LibraryRelease{
		Author:           l.GetAuthor(),
		Version:          l.GetVersion(),
		Maintainer:       l.GetMaintainer(),
		Sentence:         l.GetSentence(),
		Paragraph:        l.GetParagraph(),
		Website:          l.GetWebsite(),
		Category:         l.GetCategory(),
		Architectures:    l.GetArchitectures(),
		Types:            l.GetTypes(),
		Resources:        NewDownloadResource(l.GetResources()),
		License:          l.GetLicense(),
		ProvidesIncludes: l.GetProvidesIncludes(),
		Dependencies:     NewLibraryDependencies(l.GetDependencies()),
	}
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

func NewInstalledLibrary(l *rpc.InstalledLibrary) *InstalledLibrary {
	return &InstalledLibrary{
		Library: NewLibrary(l.GetLibrary()),
		Release: NewLibraryRelease(l.GetRelease()),
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
	if p == nil {
		return nil
	}
	propertiesMap := orderedmap.New[string, string]()
	for k, v := range p.GetProperties() {
		propertiesMap.Set(k, v)
	}
	propertiesMap.SortStableKeys(cmp.Compare)
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
	IdentificationProperties []*BoardIdentificationProperties `json:"identification_properties,omitempty"`
	BuildProperties          []string                         `json:"build_properties,omitempty"`
	DefaultProgrammerID      string                           `json:"default_programmer_id,omitempty"`
}

func NewBoardDetailsResponse(b *rpc.BoardDetailsResponse) *BoardDetailsResponse {
	if b == nil {
		return nil
	}
	buildProperties := b.GetBuildProperties()
	slices.Sort(buildProperties)
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
		IdentificationProperties: NewBoardIdentificationProperties(b.GetIdentificationProperties()),
		BuildProperties:          buildProperties,
		DefaultProgrammerID:      b.GetDefaultProgrammerId(),
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

	slices.SortFunc(res, func(a, b *Programmer) int {
		return cmp.Compare(a.Id, b.Id)
	})
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

type BoardListAllResponse struct {
	Boards []*BoardListItem `json:"boards,omitempty"`
}

func NewBoardListAllResponse(p *rpc.BoardListAllResponse) *BoardListAllResponse {
	if p == nil {
		return nil
	}
	boards := make([]*BoardListItem, len(p.GetBoards()))
	for i, v := range p.GetBoards() {
		boards[i] = NewBoardListItem(v)
	}
	return &BoardListAllResponse{Boards: boards}
}

type BoardListItem struct {
	Name     string    `json:"name,omitempty"`
	Fqbn     string    `json:"fqbn,omitempty"`
	IsHidden bool      `json:"is_hidden,omitempty"`
	Platform *Platform `json:"platform,omitempty"`
}

func NewBoardListItems(b []*rpc.BoardListItem) []*BoardListItem {
	if b == nil {
		return nil
	}
	res := make([]*BoardListItem, len(b))
	for i, v := range b {
		res[i] = NewBoardListItem(v)
	}
	return res
}

func NewBoardListItem(b *rpc.BoardListItem) *BoardListItem {
	if b == nil {
		return nil
	}
	return &BoardListItem{
		Name:     b.GetName(),
		Fqbn:     b.GetFqbn(),
		IsHidden: b.GetIsHidden(),
		Platform: NewPlatform(b.GetPlatform()),
	}
}

type Platform struct {
	Metadata *PlatformMetadata `json:"metadata,omitempty"`
	Release  *PlatformRelease  `json:"release,omitempty"`
}

func NewPlatform(p *rpc.Platform) *Platform {
	if p == nil {
		return nil
	}
	return &Platform{
		Metadata: NewPlatformMetadata(p.GetMetadata()),
		Release:  NewPlatformRelease(p.GetRelease()),
	}
}

type PlatformMetadata struct {
	Id                string `json:"id,omitempty"`
	Maintainer        string `json:"maintainer,omitempty"`
	Website           string `json:"website,omitempty"`
	Email             string `json:"email,omitempty"`
	ManuallyInstalled bool   `json:"manually_installed,omitempty"`
	Deprecated        bool   `json:"deprecated,omitempty"`
	Indexed           bool   `json:"indexed,omitempty"`
}

func NewPlatformMetadata(p *rpc.PlatformMetadata) *PlatformMetadata {
	if p == nil {
		return nil
	}
	return &PlatformMetadata{
		Id:                p.GetId(),
		Maintainer:        p.GetMaintainer(),
		Website:           p.GetWebsite(),
		Email:             p.GetEmail(),
		ManuallyInstalled: p.GetManuallyInstalled(),
		Deprecated:        p.GetDeprecated(),
		Indexed:           p.GetIndexed(),
	}
}

type DetectedPort struct {
	MatchingBoards []*BoardListItem `json:"matching_boards,omitempty"`
	Port           *Port            `json:"port,omitempty"`
}

func NewDetectedPorts(p []*rpc.DetectedPort) []*DetectedPort {
	if p == nil {
		return nil
	}
	res := make([]*DetectedPort, len(p))
	for i, v := range p {
		res[i] = NewDetectedPort(v)
	}
	return res
}

func NewDetectedPort(p *rpc.DetectedPort) *DetectedPort {
	if p == nil {
		return nil
	}
	return &DetectedPort{
		MatchingBoards: NewBoardListItems(p.GetMatchingBoards()),
		Port:           NewPort(p.GetPort()),
	}
}

type LibraryResolveDependenciesResponse struct {
	Dependencies []*LibraryDependencyStatus `json:"dependencies,omitempty"`
}

func NewLibraryResolveDependenciesResponse(l *rpc.LibraryResolveDependenciesResponse) *LibraryResolveDependenciesResponse {
	if l == nil {
		return nil
	}
	dependencies := make([]*LibraryDependencyStatus, len(l.GetDependencies()))
	for i, v := range l.GetDependencies() {
		dependencies[i] = NewLibraryDependencyStatus(v)
	}
	return &LibraryResolveDependenciesResponse{Dependencies: dependencies}
}

type LibraryDependencyStatus struct {
	Name             string `json:"name,omitempty"`
	VersionRequired  string `json:"version_required,omitempty"`
	VersionInstalled string `json:"version_installed,omitempty"`
}

func NewLibraryDependencyStatus(l *rpc.LibraryDependencyStatus) *LibraryDependencyStatus {
	if l == nil {
		return nil
	}
	return &LibraryDependencyStatus{
		Name:             l.GetName(),
		VersionRequired:  l.GetVersionRequired(),
		VersionInstalled: l.GetVersionInstalled(),
	}
}

type LibrarySearchStatus string

const (
	LibrarySearchStatusFailed  LibrarySearchStatus = "failed"
	LibrarySearchStatusSuccess LibrarySearchStatus = "success"
)

func NewLibrarySearchStatus(r rpc.LibrarySearchStatus) LibrarySearchStatus {
	switch r {
	case rpc.LibrarySearchStatus_LIBRARY_SEARCH_STATUS_FAILED:
		return LibrarySearchStatusFailed
	case rpc.LibrarySearchStatus_LIBRARY_SEARCH_STATUS_SUCCESS:
		return LibrarySearchStatusSuccess
	}
	return LibrarySearchStatusFailed
}

type LibrarySearchResponse struct {
	Libraries []*SearchedLibrary  `json:"libraries,omitempty"`
	Status    LibrarySearchStatus `json:"status,omitempty"`
}

func NewLibrarySearchResponse(l *rpc.LibrarySearchResponse) *LibrarySearchResponse {
	if l == nil {
		return nil
	}

	searchedLibraries := make([]*SearchedLibrary, len(l.GetLibraries()))
	for i, v := range l.GetLibraries() {
		searchedLibraries[i] = NewSearchedLibrary(v)
	}

	return &LibrarySearchResponse{
		Libraries: searchedLibraries,
		Status:    NewLibrarySearchStatus(l.GetStatus()),
	}
}

type SearchedLibrary struct {
	Name              string                                           `json:"name,omitempty"`
	Releases          orderedmap.Map[*semver.Version, *LibraryRelease] `json:"releases,omitempty"`
	Latest            *LibraryRelease                                  `json:"latest,omitempty"`
	AvailableVersions []string                                         `json:"available_versions,omitempty"`
}

func NewSearchedLibrary(l *rpc.SearchedLibrary) *SearchedLibrary {
	if l == nil {
		return nil
	}
	releasesMap := orderedmap.NewWithConversionFunc[*semver.Version, *LibraryRelease, string]((*semver.Version).String)
	for k, v := range l.GetReleases() {
		releasesMap.Set(semver.MustParse(k), NewLibraryRelease(v))
	}
	releasesMap.SortKeys((*semver.Version).CompareTo)
	return &SearchedLibrary{
		Name:              l.GetName(),
		Releases:          releasesMap,
		Latest:            NewLibraryRelease(l.GetLatest()),
		AvailableVersions: l.GetAvailableVersions(),
	}
}

type MonitorPortSettingDescriptor struct {
	SettingId  string   `json:"setting_id,omitempty"`
	Label      string   `json:"label,omitempty"`
	Type       string   `json:"type,omitempty"`
	EnumValues []string `json:"enum_values,omitempty"`
	Value      string   `json:"value,omitempty"`
}

func NewMonitorPortSettingDescriptor(m *rpc.MonitorPortSettingDescriptor) *MonitorPortSettingDescriptor {
	if m == nil {
		return nil
	}
	return &MonitorPortSettingDescriptor{
		SettingId:  m.GetSettingId(),
		Label:      m.GetLabel(),
		Type:       m.GetType(),
		EnumValues: m.GetEnumValues(),
		Value:      m.GetValue(),
	}
}

type BuilderResult struct {
	BuildPath              string                      `json:"build_path,omitempty"`
	UsedLibraries          []*Library                  `json:"used_libraries,omitempty"`
	ExecutableSectionsSize []*ExecutableSectionSize    `json:"executable_sections_size,omitempty"`
	BoardPlatform          *InstalledPlatformReference `json:"board_platform,omitempty"`
	BuildPlatform          *InstalledPlatformReference `json:"build_platform,omitempty"`
	BuildProperties        []string                    `json:"build_properties,omitempty"`
	Diagnostics            []*CompileDiagnostic        `json:"diagnostics,omitempty"`
}

func NewBuilderResult(c *rpc.BuilderResult) *BuilderResult {
	if c == nil {
		return nil
	}
	usedLibs := make([]*Library, len(c.GetUsedLibraries()))
	for i, v := range c.GetUsedLibraries() {
		usedLibs[i] = NewLibrary(v)
	}
	executableSectionsSizes := make([]*ExecutableSectionSize, len(c.GetExecutableSectionsSize()))
	for i, v := range c.GetExecutableSectionsSize() {
		executableSectionsSizes[i] = NewExecutableSectionSize(v)
	}

	return &BuilderResult{
		BuildPath:              c.GetBuildPath(),
		UsedLibraries:          usedLibs,
		ExecutableSectionsSize: executableSectionsSizes,
		BoardPlatform:          NewInstalledPlatformReference(c.GetBoardPlatform()),
		BuildPlatform:          NewInstalledPlatformReference(c.GetBuildPlatform()),
		BuildProperties:        c.GetBuildProperties(),
		Diagnostics:            NewCompileDiagnostics(c.GetDiagnostics()),
	}
}

type ExecutableSectionSize struct {
	Name    string `json:"name,omitempty"`
	Size    int64  `json:"size,omitempty"`
	MaxSize int64  `json:"max_size,omitempty"`
}

func NewExecutableSectionSize(s *rpc.ExecutableSectionSize) *ExecutableSectionSize {
	if s == nil {
		return nil
	}
	return &ExecutableSectionSize{
		Name:    s.GetName(),
		Size:    s.GetSize(),
		MaxSize: s.GetMaxSize(),
	}
}

type InstalledPlatformReference struct {
	Id         string `json:"id,omitempty"`
	Version    string `json:"version,omitempty"`
	InstallDir string `json:"install_dir,omitempty"`
	PackageUrl string `json:"package_url,omitempty"`
}

func NewInstalledPlatformReference(r *rpc.InstalledPlatformReference) *InstalledPlatformReference {
	if r == nil {
		return nil
	}
	return &InstalledPlatformReference{
		Id:         r.GetId(),
		Version:    r.GetVersion(),
		InstallDir: r.GetInstallDir(),
		PackageUrl: r.GetPackageUrl(),
	}
}

type BoardListWatchResponse struct {
	EventType string        `json:"event_type,omitempty"`
	Port      *DetectedPort `json:"port,omitempty"`
	Error     string        `json:"error,omitempty"`
}

func NewBoardListWatchResponse(r *rpc.BoardListWatchResponse) *BoardListWatchResponse {
	if r == nil {
		return nil
	}
	return &BoardListWatchResponse{
		EventType: r.GetEventType(),
		Port:      NewDetectedPort(r.GetPort()),
		Error:     r.GetError(),
	}
}

type CompileDiagnostic struct {
	Severity string                      `json:"severity,omitempty"`
	Message  string                      `json:"message,omitempty"`
	File     string                      `json:"file,omitempty"`
	Line     int64                       `json:"line,omitempty"`
	Column   int64                       `json:"column,omitempty"`
	Context  []*CompileDiagnosticContext `json:"context,omitempty"`
	Notes    []*CompileDiagnosticNote    `json:"notes,omitempty"`
}

func NewCompileDiagnostics(cd []*rpc.CompileDiagnostic) []*CompileDiagnostic {
	return f.Map(cd, NewCompileDiagnostic)
}

func NewCompileDiagnostic(cd *rpc.CompileDiagnostic) *CompileDiagnostic {
	return &CompileDiagnostic{
		Severity: cd.GetSeverity(),
		Message:  cd.GetMessage(),
		File:     cd.GetFile(),
		Line:     cd.GetLine(),
		Column:   cd.GetColumn(),
		Context:  f.Map(cd.GetContext(), NewCompileDiagnosticContext),
		Notes:    f.Map(cd.GetNotes(), NewCompileDiagnosticNote),
	}
}

type CompileDiagnosticContext struct {
	Message string `json:"message,omitempty"`
	File    string `json:"file,omitempty"`
	Line    int64  `json:"line,omitempty"`
	Column  int64  `json:"column,omitempty"`
}

func NewCompileDiagnosticContext(cdc *rpc.CompileDiagnosticContext) *CompileDiagnosticContext {
	return &CompileDiagnosticContext{
		Message: cdc.GetMessage(),
		File:    cdc.GetFile(),
		Line:    cdc.GetLine(),
		Column:  cdc.GetColumn(),
	}
}

type CompileDiagnosticNote struct {
	Message string `json:"message,omitempty"`
	File    string `json:"file,omitempty"`
	Line    int64  `json:"line,omitempty"`
	Column  int64  `json:"column,omitempty"`
}

func NewCompileDiagnosticNote(cdn *rpc.CompileDiagnosticNote) *CompileDiagnosticNote {
	return &CompileDiagnosticNote{
		Message: cdn.GetMessage(),
		File:    cdn.GetFile(),
		Line:    cdn.GetLine(),
		Column:  cdn.GetColumn(),
	}
}

type IsDebugSupportedResponse struct {
	DebuggingSupported bool   `json:"debugging_supported"`
	DebugFQBN          string `json:"debug_fqbn,omitempty"`
}

func NewIsDebugSupportedResponse(resp *rpc.IsDebugSupportedResponse) *IsDebugSupportedResponse {
	return &IsDebugSupportedResponse{
		DebuggingSupported: resp.GetDebuggingSupported(),
		DebugFQBN:          resp.GetDebugFqbn(),
	}
}

type UpdateIndexResponse_ResultResult struct {
	UpdatedIndexes []*IndexUpdateReportResult `json:"updated_indexes,omitempty"`
}

func NewUpdateIndexResponse_ResultResult(resp *rpc.UpdateIndexResponse_Result) *UpdateIndexResponse_ResultResult {
	return &UpdateIndexResponse_ResultResult{
		UpdatedIndexes: f.Map(resp.GetUpdatedIndexes(), NewIndexUpdateReportResult),
	}
}

type UpdateLibrariesIndexResponse_ResultResult struct {
	LibrariesIndex *IndexUpdateReportResult `json:"libraries_index"`
}

func NewUpdateLibrariesIndexResponse_ResultResult(resp *rpc.UpdateLibrariesIndexResponse_Result) *UpdateLibrariesIndexResponse_ResultResult {
	return &UpdateLibrariesIndexResponse_ResultResult{
		LibrariesIndex: NewIndexUpdateReportResult(resp.GetLibrariesIndex()),
	}
}

type IndexUpdateReportResult struct {
	IndexURL string                   `json:"index_url"`
	Status   IndexUpdateReport_Status `json:"status"`
}

func NewIndexUpdateReportResult(resp *rpc.IndexUpdateReport) *IndexUpdateReportResult {
	return &IndexUpdateReportResult{
		IndexURL: resp.GetIndexUrl(),
		Status:   NewIndexUpdateReport_Status(resp.GetStatus()),
	}
}

type IndexUpdateReport_Status string

const (
	IndexUpdateReport_StatusUnspecified     IndexUpdateReport_Status = "unspecified"
	IndexUpdateReport_StatusAlreadyUpToDate IndexUpdateReport_Status = "already-up-to-date"
	IndexUpdateReport_StatusFailed          IndexUpdateReport_Status = "failed"
	IndexUpdateReport_StatusSkipped         IndexUpdateReport_Status = "skipped"
	IndexUpdateReport_StatusUpdated         IndexUpdateReport_Status = "updated"
)

func NewIndexUpdateReport_Status(r rpc.IndexUpdateReport_Status) IndexUpdateReport_Status {
	switch r {
	case rpc.IndexUpdateReport_STATUS_UNSPECIFIED:
		return IndexUpdateReport_StatusUnspecified
	case rpc.IndexUpdateReport_STATUS_UPDATED:
		return IndexUpdateReport_StatusUpdated
	case rpc.IndexUpdateReport_STATUS_ALREADY_UP_TO_DATE:
		return IndexUpdateReport_StatusAlreadyUpToDate
	case rpc.IndexUpdateReport_STATUS_FAILED:
		return IndexUpdateReport_StatusFailed
	case rpc.IndexUpdateReport_STATUS_SKIPPED:
		return IndexUpdateReport_StatusSkipped
	default:
		return IndexUpdateReport_StatusUnspecified
	}
}
