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

package arguments

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ShowProperties represents the --show-properties flag.
type ShowProperties struct {
	arg string
}

// ShowPropertiesMode represents the possible values of the --show-properties flag.
type ShowPropertiesMode int

const (
	// ShowPropertiesDisabled means that the --show-properties flag has not been used
	ShowPropertiesDisabled ShowPropertiesMode = iota
	// ShowPropertiesUnexpanded means that the --show-properties flag has been used without a value or with the value "unexpanded"
	ShowPropertiesUnexpanded
	// ShowPropertiesExpanded means that the --show-properties flag has been used with the value "expanded"
	ShowPropertiesExpanded
)

// Get returns the corresponding ShowProperties value.
func (p *ShowProperties) Get() (ShowPropertiesMode, error) {
	switch p.arg {
	case "disabled":
		return ShowPropertiesDisabled, nil
	case "unexpanded":
		return ShowPropertiesUnexpanded, nil
	case "expanded":
		return ShowPropertiesExpanded, nil
	default:
		return ShowPropertiesDisabled, fmt.Errorf(tr("invalid option '%s'.", p.arg))
	}
}

// AddToCommand adds the --show-properties flag to the specified command.
func (p *ShowProperties) AddToCommand(command *cobra.Command) {
	command.Flags().StringVar(&p.arg,
		"show-properties", "disabled",
		tr(`Show build properties. The properties are expanded, use "--show-properties=unexpanded" if you want them exactly as they are defined.`),
	)
	command.Flags().Lookup("show-properties").NoOptDefVal = "expanded" // default if the flag is present with no value
}
