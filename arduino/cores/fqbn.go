/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
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
 * Copyright 2017-2018 ARDUINO AG (http://www.arduino.cc/)
 */

package cores

import (
	"fmt"
	"sort"
	"strings"

	properties "github.com/arduino/go-properties-map"
)

// FQBN represents a Board with a specific configuration
type FQBN struct {
	Package      string
	PlatformArch string
	BoardID      string
	Configs      properties.Map
}

// ParseFQBN extract an FQBN object from the input string
func ParseFQBN(fqbnIn string) (*FQBN, error) {
	// Split fqbn
	fqbnParts := strings.Split(fqbnIn, ":")
	if len(fqbnParts) < 3 || len(fqbnParts) > 4 {
		return nil, fmt.Errorf("invalid fqbn: %s", fqbnIn)
	}

	fqbn := &FQBN{
		Package:      fqbnParts[0],
		PlatformArch: fqbnParts[1],
		BoardID:      fqbnParts[2],
	}
	if fqbn.BoardID == "" {
		return nil, fmt.Errorf("invalid fqbn: empty board identifier")
	}
	if len(fqbnParts) > 3 {
		fqbn.Configs = properties.Map{}
		for _, pair := range strings.Split(fqbnParts[3], ",") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid fqbn config: %s", pair)
			}
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if k == "" {
				return nil, fmt.Errorf("invalid fqbn config: %s", pair)
			}
			fqbn.Configs[k] = v
		}
	}
	return fqbn, nil
}

func (fqbn *FQBN) String() string {
	res := fmt.Sprintf("%s:%s:%s", fqbn.Package, fqbn.PlatformArch, fqbn.BoardID)
	if fqbn.Configs != nil {
		sep := ":"
		keys := fqbn.Configs.Keys()
		sort.Strings(keys)
		for _, k := range keys {
			res += sep + k + "=" + fqbn.Configs[k]
			sep = ","
		}
	}
	return res
}
