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

package builderclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ShowFilesPath computes a request path to the show action of files.
func ShowFilesPath(path string) string {
	return fmt.Sprintf("/builder/v1/files/%s", path)
}

// ShowFiles provides the content of a file.
func (c *Client) ShowFiles(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewShowFilesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowFilesRequest create the request corresponding to the show action endpoint of the files resource.
func (c *Client) NewShowFilesRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}
