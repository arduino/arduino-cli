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

// ListLibrariesPath computes a request path to the list action of libraries.
func ListLibrariesPath() string {
	return fmt.Sprintf("/builder/v1/libraries")
}

// ListLibraries provides a list of all the latest versions of the libraries supported by Arduino Create. Doesn't require any authentication.
func (c *Client) ListLibraries(ctx context.Context, path string, maintainer *string, type1 *string, withoutType *string) (*http.Response, error) {
	req, err := c.NewListLibrariesRequest(ctx, path, maintainer, type1, withoutType)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewListLibrariesRequest create the request corresponding to the list action endpoint of the libraries resource.
func (c *Client) NewListLibrariesRequest(ctx context.Context, path string, maintainer *string, type1 *string, withoutType *string) (*http.Request, error) {
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
	if withoutType != nil {
		values.Set("without_type", *withoutType)
	}
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// ShowLibrariesPath computes a request path to the show action of libraries.
func ShowLibrariesPath(id string) string {
	return fmt.Sprintf("/builder/v1/libraries/%s", id)
}

// ShowLibraries provides the library identified by the :id and :pid param. Doesn't require authentication. Also contains a list of other versions of the library
func (c *Client) ShowLibraries(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewShowLibrariesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowLibrariesRequest create the request corresponding to the show action endpoint of the libraries resource.
func (c *Client) NewShowLibrariesRequest(ctx context.Context, path string) (*http.Request, error) {
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
