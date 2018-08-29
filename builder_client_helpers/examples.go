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

// ListExamplesPath computes a request path to the list action of examples.
func ListExamplesPath() string {
	return fmt.Sprintf("/builder/v1/examples")
}

// ListExamples provides a list of all the builtin examples
func (c *Client) ListExamples(ctx context.Context, path string, maintainer *string, type1 *string) (*http.Response, error) {
	req, err := c.NewListExamplesRequest(ctx, path, maintainer, type1)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewListExamplesRequest create the request corresponding to the list action endpoint of the examples resource.
func (c *Client) NewListExamplesRequest(ctx context.Context, path string, maintainer *string, type1 *string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	values := u.Query()
	if maintainer != nil {
		values.Set("maintainer", *maintainer)
	}
	if type1 != nil {
		values.Set("type", *type1)
	}
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}
