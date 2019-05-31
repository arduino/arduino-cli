/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package test

import (
	"fmt"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestI18NSyntax(t *testing.T) {
	require.Equal(t, "Do you want to remove %[1]v?\nIf you do so you won't be able to use %[1]v any more.", i18n.FromJavaToGoSyntax("Do you want to remove {0}?\nIf you do so you won't be able to use {0} any more."))
	require.Equal(t, "A file named \"%[1]v\" already exists in \"%[2]v\"", i18n.FromJavaToGoSyntax("A file named \"{0}\" already exists in \"{1}\""))
	require.Equal(t, "Board %[1]v:%[2]v:%[3]v doesn't define a 'build.board' preference. Auto-set to: %[4]v", i18n.FromJavaToGoSyntax("Board {0}:{1}:{2} doesn''t define a ''build.board'' preference. Auto-set to: {3}"))

	require.Equal(t, "22 11\n", fmt.Sprintf("%[2]d %[1]d\n", 11, 22))
	require.Equal(t, "d c b a", i18n.Format("{3} {2} {1} {0}", "a", "b", "c", "d"))

	require.Equal(t, "e d b a", i18n.Format("{4} {3} {1} {0}", "a", "b", "c", "d", "e"))

	require.Equal(t, "a b", i18n.Format("{0} {1}", "a", "b", "c", "d", "e"))

	require.Equal(t, "%!v(BADINDEX) c b a", i18n.Format("{3} {2} {1} {0}", "a", "b", "c"))
	require.Equal(t, "%!v(BADINDEX) %!v(BADINDEX) %!v(BADINDEX) %!v(BADINDEX)", i18n.Format("{3} {2} {1} {0}"))

	require.Equal(t, "I'm %[1]v%% sure", i18n.FromJavaToGoSyntax("I'm {0}%% sure"))
	require.Equal(t, "I'm 100% sure", i18n.Format("I'm {0}%% sure", 100))

	require.Equal(t, "Either in [1] or in [2]", i18n.Format("Either in [{0}] or in [{1}]", 1, 2))

	require.Equal(t, "Using library a at version b in folder: c ", i18n.Format("Using library {0} at version {1} in folder: {2} {3}", "a", "b", "c", ""))

	require.Equal(t, "Missing 'a' from library in b", i18n.Format("Missing '{0}' from library in {1}", "a", "b"))
}

func TestI18NInheritance(t *testing.T) {
	var logger i18n.Logger
	logger = i18n.HumanLogger{}
	logger.Println(constants.LOG_LEVEL_INFO, "good {0} {1}", "morning", "vietnam!")

	logger = i18n.MachineLogger{}
	logger.Println(constants.LOG_LEVEL_INFO, "good {0} {1}", "morning", "vietnam!")
}
