// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package configuration

func (settings *Settings) BoardManagerAdditionalUrls() []string {
	if urls, ok, _ := settings.GetStringSliceOk("board_manager.additional_urls"); ok {
		return urls
	}
	return settings.Defaults.GetStringSlice("board_manager.additional_urls")
}

func (settings *Settings) BoardManagerEnableUnsafeInstall() bool {
	if v, ok, _ := settings.GetBoolOk("board_manager.enable_unsafe_install"); ok {
		return v
	}
	return settings.Defaults.GetBool("board_manager.enable_unsafe_install")
}
