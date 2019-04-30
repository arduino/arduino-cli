/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package lib

import (
	"context"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
)

func LibrarySearch(ctx context.Context, req *rpc.LibrarySearchReq) (*rpc.LibrarySearchResp, error) {

	lm := commands.GetLibraryManager(req)

	res := []*rpc.SearchLibraryOutput{}
	release, index string

	for _, lib := range lm.Index.Libraries {
		if strings.Contains(strings.ToLower(lib.Name), req.GetQuery()) {
			for rel := range lib.Releases {
				release += rel;
			}

			for idx := range &lib.Libraries {
				index = 
			}
			res = append(res, &rpc.SearchLibraryOutput{
				Name:     lib.Name,
				Releases: release,
				Latest:   lib.Latest.String(),
				Index:    index,
			})
		}
	}

	if req.GetNames() {
		for _, lib := range res.Libraries {
			formatter.Print(lib.Name)
		}
	} else {
		if len(res.Libraries) == 0 {
			formatter.Print(fmt.Sprintf("No library found matching `%s` search query", req.GetQuery()))
		} else {
			formatter.Print(res)
		}
	}

	return &rpc.LibrarySearchResp{}, nil
}
