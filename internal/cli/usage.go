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

package cli

import "github.com/arduino/arduino-cli/internal/i18n"

func getUsageTemplate() string {
	// Force i18n to generate translation strings
	_ = i18n.Tr("Usage:")
	_ = i18n.Tr("Aliases:")
	_ = i18n.Tr("Examples:")
	_ = i18n.Tr("Available Commands:")
	_ = i18n.Tr("Flags:")
	_ = i18n.Tr("Global Flags:")
	_ = i18n.Tr("Additional help topics:")
	_ = i18n.Tr("Use %s for more information about a command.")

	return `{{tr "Usage:"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{tr "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{tr "Examples:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

{{tr "Available Commands:"}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{tr "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{tr "Global Flags:"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{tr "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

{{tr "Use %s for more information about a command." (printf "%s %s" .CommandPath "[command] --help" | printf "%q")}}{{end}}
`
}
