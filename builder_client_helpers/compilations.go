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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package builderClient

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

// Start a compilation for the given user and saves the request (but not the generated files) on the database. requires authentication. Can return PreconditionFailed if the user has reached their maximum number of compilations per day. If the compilation failed it returns UnprocessableEntity
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
