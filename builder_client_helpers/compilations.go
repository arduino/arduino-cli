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
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// StartCompilationsPath computes a request path to the start action of compilations.
func StartCompilationsPath() string {
	return fmt.Sprintf("/builder/v1/compile")
}

// Start a compilation for the given user and saves the request (but not the generated files) on the database.
// requires authentication.
// Can return PreconditionFailed if the user has reached their maximum number of compilations per day.
// If the compilation failed it returns UnprocessableEntity
func (c *Client) StartCompilations(ctx context.Context, path string, payload *Compilation) (*http.Response, error) {
	req, err := c.NewStartCompilationsRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewStartCompilationsRequest create the request corresponding to the start action endpoint of the compilations resource.
func (c *Client) NewStartCompilationsRequest(ctx context.Context, path string, payload *Compilation) (*http.Request, error) {
	var body bytes.Buffer
	err := c.Encoder.Encode(payload, &body, "*/*")
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %s", err)
	}
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("POST", u.String(), &body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")
	if c.Oauth2Signer != nil {
		c.Oauth2Signer.Sign(req)
	}
	return req, nil
}
