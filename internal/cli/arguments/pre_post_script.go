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
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// PrePostScriptsFlags contains flags data used by the core install and the upgrade command
// This is useful so all flags used by commands that need
// this information are consistent with each other.
type PrePostScriptsFlags struct {
	runPostInstall   bool // force the execution of installation scripts
	skipPostInstall  bool // skip the execution of installation scripts
	runPreUninstall  bool // force the execution of pre uninstall scripts
	skipPreUninstall bool // skip the execution of pre uninstall scripts
}

// AddToCommand adds flags that can be used to force running or skipping
// of post installation scripts
func (p *PrePostScriptsFlags) AddToCommand(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&p.runPostInstall, "run-post-install", false, i18n.Tr("Force run of post-install scripts (if the CLI is not running interactively)."))
	cmd.Flags().BoolVar(&p.skipPostInstall, "skip-post-install", false, i18n.Tr("Force skip of post-install scripts (if the CLI is running interactively)."))
	cmd.Flags().BoolVar(&p.runPreUninstall, "run-pre-uninstall", false, i18n.Tr("Force run of pre-uninstall scripts (if the CLI is not running interactively)."))
	cmd.Flags().BoolVar(&p.skipPreUninstall, "skip-pre-uninstall", false, i18n.Tr("Force skip of pre-uninstall scripts (if the CLI is running interactively)."))
}

// GetRunPostInstall returns the run-post-install flag value
func (p *PrePostScriptsFlags) GetRunPostInstall() bool {
	return p.runPostInstall
}

// GetSkipPostInstall returns the skip-post-install flag value
func (p *PrePostScriptsFlags) GetSkipPostInstall() bool {
	return p.skipPostInstall
}

// GetRunPreUninstall returns the run-post-install flag value
func (p *PrePostScriptsFlags) GetRunPreUninstall() bool {
	return p.runPreUninstall
}

// GetSkipPreUninstall returns the skip-post-install flag value
func (p *PrePostScriptsFlags) GetSkipPreUninstall() bool {
	return p.skipPreUninstall
}

// DetectSkipPostInstallValue returns true if a post install script must be run
func (p *PrePostScriptsFlags) DetectSkipPostInstallValue() bool {
	if p.GetRunPostInstall() {
		logrus.Info("Will run post-install by user request")
		return false
	}
	if p.GetSkipPostInstall() {
		logrus.Info("Will skip post-install by user request")
		return true
	}

	if !feedback.IsInteractive() {
		logrus.Info("Not running from console, will skip post-install by default")
		return true
	}
	logrus.Info("Running from console, will run post-install by default")
	return false
}

// DetectSkipPreUninstallValue returns true if a post install script must be run
func (p *PrePostScriptsFlags) DetectSkipPreUninstallValue() bool {
	if p.GetRunPreUninstall() {
		logrus.Info("Will run pre-uninstall by user request")
		return false
	}
	if p.GetSkipPreUninstall() {
		logrus.Info("Will skip pre-uninstall by user request")
		return true
	}

	if !feedback.IsInteractive() {
		logrus.Info("Not running from console, will skip pre-uninstall by default")
		return true
	}
	logrus.Info("Running from console, will run pre-uninstall by default")
	return false
}
