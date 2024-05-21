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

package cores

import (
	"fmt"
	"strings"
	"sync"

	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/go-properties-orderedmap"
)

// Board represents a board loaded from an installed platform
type Board struct {
	BoardID                  string
	Properties               *properties.Map  `json:"-"`
	PlatformRelease          *PlatformRelease `json:"-"`
	configOptionsMux         sync.Mutex
	configOptions            *properties.Map
	configOptionValues       map[string]*properties.Map
	configOptionProperties   map[string]*properties.Map
	defaultConfig            *properties.Map
	identificationProperties []*properties.Map
}

// HasUsbID returns true if the board match the usb vid and pid parameters
func (b *Board) HasUsbID(reqVid, reqPid string) bool {
	vids := b.Properties.SubTree("vid")
	pids := b.Properties.SubTree("pid")
	for id, vid := range vids.AsMap() {
		if pid, ok := pids.GetOk(id); ok {
			if strings.EqualFold(vid, reqVid) && strings.EqualFold(pid, reqPid) {
				return true
			}
		}
	}
	return false
}

// Name returns the board name as defined in boards.txt properties
func (b *Board) Name() string {
	return b.Properties.Get("name")
}

// FQBN return the Fully-Qualified-Board-Name for the default configuration of this board
func (b *Board) FQBN() string {
	platform := b.PlatformRelease.Platform
	return platform.Package.Name + ":" + platform.Architecture + ":" + b.BoardID
}

// IsHidden returns true if the board is marked as hidden in the platform
func (b *Board) IsHidden() bool {
	return b.Properties.GetBoolean("hide")
}

func (b *Board) String() string {
	return b.FQBN()
}

func (b *Board) buildConfigOptionsStructures() {
	b.configOptionsMux.Lock()
	defer b.configOptionsMux.Unlock()
	if b.configOptions != nil {
		return
	}

	b.configOptions = properties.NewMap()
	allConfigs := b.Properties.SubTree("menu")
	allConfigOptions := allConfigs.FirstLevelOf()

	// Used to show the config options in the same order as the menu, defined at the begging of boards.txt
	if b.PlatformRelease.Menus != nil {
		for _, menuOption := range b.PlatformRelease.Menus.FirstLevelKeys() {
			if _, ok := allConfigOptions[menuOption]; ok {
				b.configOptions.Set(menuOption, b.PlatformRelease.Menus.Get(menuOption))
			}
		}
	}

	b.configOptionValues = map[string]*properties.Map{}
	b.configOptionProperties = map[string]*properties.Map{}
	b.defaultConfig = properties.NewMap()
	for option, optionProps := range allConfigOptions {
		b.configOptionValues[option] = properties.NewMap()
		values := optionProps.FirstLevelKeys()
		b.defaultConfig.Set(option, values[0])
		for _, value := range values {
			if label, ok := optionProps.GetOk(value); ok {
				b.configOptionValues[option].Set(value, label)
				b.configOptionProperties[option+"="+value] = optionProps.SubTree(value)
			}
		}
	}
}

// GetConfigOptions returns an OrderedMap of configuration options for this board.
// The returned map will have key and value as option id and option name, respectively.
func (b *Board) GetConfigOptions() *properties.Map {
	b.buildConfigOptionsStructures()
	return b.configOptions
}

// GetConfigOptionValues returns an OrderedMap of possible values for a specific configuratio options
// for this board. The returned map will have key and value as option value and option value name,
// respectively.
func (b *Board) GetConfigOptionValues(option string) *properties.Map {
	b.buildConfigOptionsStructures()
	return b.configOptionValues[option]
}

// GetBuildProperties returns the build properties and the build
// platform for the Board with the configuration passed as parameter.
func (b *Board) GetBuildProperties(fqbn *FQBN) (*properties.Map, error) {
	b.buildConfigOptionsStructures()

	// Override default configs with user configs
	config := b.defaultConfig.Clone()
	config.Merge(fqbn.Configs)

	// Start with board's base properties
	buildProperties := b.Properties.Clone()
	buildProperties.Set("build.fqbn", fqbn.String())
	buildProperties.Set("build.arch", strings.ToUpper(b.PlatformRelease.Platform.Architecture))

	// Add all sub-configurations one by one (a config is: option=value)
	// Check for residual invalid options...
	for option, value := range config.AsMap() {
		if option == "" {
			return nil, fmt.Errorf(i18n.Tr("invalid empty option found"))
		}
		if _, ok := b.configOptions.GetOk(option); !ok {
			return nil, fmt.Errorf(i18n.Tr("invalid option '%s'"), option)
		}
		optionsConf, ok := b.configOptionProperties[option+"="+value]
		if !ok {
			return nil, fmt.Errorf(i18n.Tr("invalid value '%[1]s' for option '%[2]s'"), value, option)
		}
		buildProperties.Merge(optionsConf)
	}

	return buildProperties, nil
}

// GeneratePropertiesForConfiguration returns the board properties for a particular
// configuration. The parameter is the latest part of the FQBN, for example if
// the full FQBN is "arduino:avr:mega:cpu=atmega2560" the config part must be
// "cpu=atmega2560".
// FIXME: deprecated, use GetBuildProperties instead
func (b *Board) GeneratePropertiesForConfiguration(config string) (*properties.Map, error) {
	fqbn, err := ParseFQBN(b.String() + ":" + config)
	if err != nil {
		return nil, fmt.Errorf(i18n.Tr("parsing fqbn: %s"), err)
	}
	return b.GetBuildProperties(fqbn)
}

// GetIdentificationProperties calculates and returns a list of properties sets
// containing the properties required to identify the board. The returned sets
// must not be changed by the caller.
func (b *Board) GetIdentificationProperties() []*properties.Map {
	if b.identificationProperties == nil {
		b.identificationProperties = b.Properties.ExtractSubIndexSets("upload_port")
	}
	return b.identificationProperties
}

// IsBoardMatchingIDProperties returns true if the board match the given
// upload port identification properties
func (b *Board) IsBoardMatchingIDProperties(query *properties.Map) bool {
	// check checks if the given set of properties p match the "query"
	check := func(p *properties.Map) bool {
		for k, v := range p.AsMap() {
			if !strings.EqualFold(query.Get(k), v) {
				return false
			}
		}
		return true
	}

	// First check the identification properties with sub index "upload_port.N.xxx"
	for _, idProps := range b.GetIdentificationProperties() {
		if check(idProps) {
			return true
		}
	}
	return false
}

// GetMonitorSettings returns the settings for the pluggable monitor of the given protocol
// and set of board properties.
func GetMonitorSettings(protocol string, boardProperties *properties.Map) *properties.Map {
	return boardProperties.SubTree("monitor_port." + protocol)
}

// IdentifyBoardConfiguration returns the configuration of the board that can be
// deduced from the given upload port identification properties
func (b *Board) IdentifyBoardConfiguration(query *properties.Map) *properties.Map {
	// check checks if the given set of properties p match the "query"
	check := func(p *properties.Map) bool {
		for k, v := range p.AsMap() {
			if !strings.EqualFold(query.Get(k), v) {
				return false
			}
		}
		return true
	}
	checkAll := func(allP []*properties.Map) bool {
		for _, p := range allP {
			if check(p) {
				return true
			}
		}
		return false
	}

	res := properties.NewMap()
	for _, option := range b.GetConfigOptions().Keys() {
		values := b.GetConfigOptionValues(option).Keys()

		for _, value := range values {
			config := option + "=" + value
			configProps := b.configOptionProperties[config]

			if checkAll(configProps.ExtractSubIndexSets("upload_port")) {
				res.Set(option, value)
			}
		}
	}
	return res
}

// GetDefaultProgrammerID returns the board's default programmer as
// defined in 'programmer.default' property of the board. If the board
// has no default programmer the empty string is returned.
func (b *Board) GetDefaultProgrammerID() string {
	return b.Properties.Get("programmer.default")
}
