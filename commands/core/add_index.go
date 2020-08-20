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

package core

import (
	"context"
	"fmt"
	"net/url"

	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/spf13/viper"
)

// AddIndex FIXMEDOC
func AddIndex(ctx context.Context, req *rpc.AddIndexReq, downloadDB commands.DownloadProgressCB) (*rpc.AddIndexResp, error) {
	originalUrls := []string{globals.DefaultIndexURL}
	originalUrls = append(originalUrls, viper.GetStringSlice("board_manager.additional_urls")...)
	urls := originalUrls

	messages := []string{}
	for _, u := range req.IndexUrl {
		URL, err := url.Parse(u)
		if err != nil {
			messages = append(messages, err.Error())
			continue
		}

		if URL.Hostname() != "downloads.arduino.cc" {
			messages = append(messages, fmt.Sprintf(
				"WARNING: A 3rd party package index %v was added. "+
					"The boards platforms from this index have not been audited for security by Arduino. "+
					"Platforms can run custom executables on your computer, so use it at your own risk.",
				u))
		}

		urls = append(urls, URL.String())
	}

	viper.Set("board_manager.additional_urls", urls)

	_, err := commands.UpdateIndex(ctx, &rpc.UpdateIndexReq{
		Instance: req.Instance,
	}, downloadDB)
	if err != nil {
		// Rollback to previous urls in case of errors
		viper.Set("board_manager.additional_urls", originalUrls)
		return nil, err
	}

	return &rpc.AddIndexResp{Messages: messages}, nil
}
