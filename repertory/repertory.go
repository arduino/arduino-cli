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

package repertory

import (
	"path/filepath"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/gofrs/uuid"
	"github.com/spf13/viper"
)

// Store is the Read Only config storage
var Store = viper.New()

var (
	Type = "yaml"
	Name = "repertory" + "." + Type
)

// Configure configures the Read Only config storage
func Init() {
	configPath := configuration.GetDefaultArduinoDataDir()
	Store.SetConfigName(Name)
	Store.SetConfigType(Type)
	Store.AddConfigPath(configPath)
	configFilePath := filepath.Join(configPath, Name)
	// Attempt to read config file
	if err := Store.ReadInConfig(); err != nil {
		// ConfigFileNotFoundError is acceptable, anything else
		// should be reported to the user
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			generateInstallationData()
			err := Store.WriteConfigAs(configFilePath)
			if err != nil {
				feedback.Errorf("Error writing repertory file: %v", err)
			}
		} else {
			feedback.Errorf("Error reading repertory file: %v", err)
		}
	}
}

func generateInstallationData() {

	installationID, err := uuid.NewV4()
	if err != nil {
		feedback.Errorf("Error generating installation.id: %v", err)
	}
	Store.Set("installation.id", installationID.String())

	installationSecret, err := uuid.NewV4()
	if err != nil {
		feedback.Errorf("Error generating installation.secret: %v", err)
	}
	Store.Set("installation.secret", installationSecret.String())
}
