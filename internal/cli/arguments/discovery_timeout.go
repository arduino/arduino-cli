// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package arguments

import (
	"time"

	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/spf13/cobra"
)

// DiscoveryTimeout is the timeout given to discoveries to detect ports.
type DiscoveryTimeout struct {
	timeout time.Duration
}

// AddToCommand adds the flags used to set fqbn to the specified Command
func (d *DiscoveryTimeout) AddToCommand(cmd *cobra.Command) {
	cmd.Flags().DurationVar(&d.timeout, "discovery-timeout", time.Second, i18n.Tr("Max time to wait for port discovery, e.g.: 30s, 1m"))
}

// Get returns the timeout
func (d *DiscoveryTimeout) Get() time.Duration {
	return d.timeout
}
