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

package inventory

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/arduino/arduino-cli/i18n"
	"github.com/gofrs/uuid"
	"github.com/spf13/viper"
)

// Store is the Read Only config storage
var Store = viper.New()

var (
	// Type is the inventory file type
	Type = "yaml"
	// Name is the inventory file Name with Type as extension
	Name           = "inventory" + "." + Type
	tr             = i18n.Tr
	configFilePath string
)

// Init configures the Read Only config storage
func Init(configPath string) error {
	configFilePath = filepath.Join(configPath, Name)
	Store.SetConfigName(Name)
	Store.SetConfigType(Type)
	Store.AddConfigPath(configPath)
	// Attempt to read config file
	if err := Store.ReadInConfig(); err != nil {
		// ConfigFileNotFoundError is acceptable, anything else
		// should be reported to the user
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := generateInstallationData(); err != nil {
				return err
			}
			if err := WriteStore(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf(tr("reading inventory file: %w"), err)
		}
	}

	return nil
}

func generateInstallationData() error {
	installationID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf(tr("generating installation.id: %w"), err)
	}
	Store.Set("installation.id", installationID.String())

	installationSecret, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf(tr("generating installation.secret: %w"), err)
	}
	Store.Set("installation.secret", installationSecret.String())
	return nil
}

var writeStoreMux sync.Mutex

// WriteStore writes the current information from Store to configFilePath.
// Returns err if it fails.
func WriteStore() error {
	writeStoreMux.Lock()
	defer writeStoreMux.Unlock()

	configPath := filepath.Dir(configFilePath)

	// Create config dir if not present,
	// MkdirAll will retrun no error if the path already exists
	if err := os.MkdirAll(configPath, os.FileMode(0755)); err != nil {
		return fmt.Errorf(tr("invalid path creating config dir: %[1]s error: %[2]w"), configPath, err)
	}

	// Create file if not present
	err := Store.WriteConfigAs(configFilePath)
	if err != nil {
		return fmt.Errorf(tr("invalid path writing inventory file: %[1]s error: %[2]w"), configFilePath, err)
	}

	return nil
}
