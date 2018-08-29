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

// AddPinnedLibrariesPath computes a request path to the add action of pinnedLibraries.
func AddPinnedLibrariesPath(id string) string {
	param0 := id

	return fmt.Sprintf("/builder/pinned/%s", param0)
}

//AddPinnedLibraries adds a new library to the list of libraries pinned by the user
func (c *Client) AddPinnedLibraries(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewAddPinnedLibrariesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewAddPinnedLibrariesRequest create the request corresponding to the add action endpoint of the pinnedLibraries resource.
func (c *Client) NewAddPinnedLibrariesRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("PUT", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.Oauth2Signer != nil {
		c.Oauth2Signer.Sign(req)
	}
	return req, nil
}

// ListPinnedLibrariesPath computes a request path to the list action of pinnedLibraries.
func ListPinnedLibrariesPath() string {
	return fmt.Sprintf("/builder/pinned")
}

//ListPinnedLibraries provides a list of all the libraries pinned by the user
func (c *Client) ListPinnedLibraries(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewListPinnedLibrariesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewListPinnedLibrariesRequest create the request corresponding to the list action endpoint of the pinnedLibraries resource.
func (c *Client) NewListPinnedLibrariesRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.Oauth2Signer != nil {
		c.Oauth2Signer.Sign(req)
	}
	return req, nil
}

// RemovePinnedLibrariesPath computes a request path to the remove action of pinnedLibraries.
func RemovePinnedLibrariesPath(id string) string {
	param0 := id

	return fmt.Sprintf("/builder/pinned/%s", param0)
}

//RemovePinnedLibraries removes a library to the list of libraries pinned by the user
func (c *Client) RemovePinnedLibraries(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewRemovePinnedLibrariesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewRemovePinnedLibrariesRequest create the request corresponding to the remove action endpoint of the pinnedLibraries resource.
func (c *Client) NewRemovePinnedLibrariesRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.Oauth2Signer != nil {
		c.Oauth2Signer.Sign(req)
	}
	return req, nil
}
