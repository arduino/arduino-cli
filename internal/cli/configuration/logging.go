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

import "github.com/arduino/go-paths-helper"

func (s *Settings) LoggingLevel() string {
	if l, ok, _ := s.GetStringOk("logging.level"); ok {
		return l
	}
	return s.Defaults.GetString("logging.level")
}

func (s *Settings) LoggingFormat() string {
	if l, ok, _ := s.GetStringOk("logging.format"); ok {
		return l
	}
	return s.Defaults.GetString("logging.format")
}

func (s *Settings) LoggingFile() *paths.Path {
	if l, ok, _ := s.GetStringOk("logging.file"); ok && l != "" {
		return paths.New(l)
	}
	if l, ok, _ := s.Defaults.GetStringOk("logging.file"); ok && l != "" {
		return paths.New(l)
	}
	return nil
}
