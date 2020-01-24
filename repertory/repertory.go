package repertory

import (
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/gofrs/uuid"
	"github.com/spf13/viper"
	"os"
)

// Store is the Read Only config storage
var Store = viper.New()

// Configure configures the Read Only config storage
func Init() {
	configPath := configuration.GetDefaultArduinoDataDir()
	Store.SetConfigType("yaml")
	Store.SetConfigName("repertory.yaml")
	Store.AddConfigPath(configPath)
	// Attempt to read config file
	if err := Store.ReadInConfig(); err != nil {
		// ConfigFileNotFoundError is acceptable, anything else
		// should be reported to the user
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// FIXME: how should I treat this error?
			installationID, _ := uuid.NewV4()
			Store.SetDefault("installation.id", installationID.String())
			installationSecret, _ := uuid.NewV4()
			Store.SetDefault("installation.secret", installationSecret.String())
			if err = Store.SafeWriteConfigAs(configPath); err != nil {
				if os.IsNotExist(err) {
					err = Store.WriteConfigAs(configPath)
				}
			}
		} else {
			feedback.Errorf("Error reading repertory file: %v", err)

		}
	}
}
