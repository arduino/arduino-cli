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

package createclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CreateSketchesPath computes a request path to the create action of sketches.
func CreateSketchesPath() string {
	return fmt.Sprintf("/create/v1/sketches")
}

const (
	devURL  = "api-dev.arduino.cc"
	prodURL = "api.arduino.cc"
	prod    = true
)

// CreateSketches Adds a new sketch.
func (c *Client) CreateSketches(ctx context.Context, path string, payload *Sketch, authorization string) (*http.Response, error) {
	req, err := c.NewCreateSketchesRequest(ctx, path, payload, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewCreateSketchesRequest create the request corresponding to the create action endpoint of the sketches resource.
func (c *Client) NewCreateSketchesRequest(ctx context.Context, path string, payload *Sketch, authorization string) (*http.Request, error) {
	var body bytes.Buffer
	err := c.Encoder.Encode(payload, &body, "*/*")
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %s", err)
	}
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	if prod {
		path = prodURL + path
	} else {
		path = devURL + path
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("PUT", u.String(), &body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")

	header.Set("Authorization", authorization)

	return req, nil
}

// DeleteSketchesPath computes a request path to the delete action of sketches.
func DeleteSketchesPath(id string) string {
	return fmt.Sprintf("/create/v1/sketches/%s", id)
}

// DeleteSketches Removes the sketch identified by the :id param.
func (c *Client) DeleteSketches(ctx context.Context, path string, authorization string) (*http.Response, error) {
	req, err := c.NewDeleteSketchesRequest(ctx, path, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewDeleteSketchesRequest create the request corresponding to the delete action endpoint of the sketches resource.
func (c *Client) NewDeleteSketchesRequest(ctx context.Context, path string, authorization string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	if prod {
		path = prodURL + path
	} else {
		path = devURL + path
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return nil, err
	}
	header := req.Header

	header.Set("Authorization", authorization)

	return req, nil
}

// EditSketchesPath computes a request path to the edit action of sketches.
func EditSketchesPath(id string) string {
	param0 := id

	return fmt.Sprintf("/create/v1/sketches/%s", param0)
}

// EditSketches Modifies the sketch identified by the :id param.
// If a file has a valid data field, it will be modified too.
func (c *Client) EditSketches(ctx context.Context, path string, payload *Sketch, authorization string) (*http.Response, error) {
	req, err := c.NewEditSketchesRequest(ctx, path, payload, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewEditSketchesRequest create the request corresponding to the edit action endpoint of the sketches resource.
func (c *Client) NewEditSketchesRequest(ctx context.Context, path string, payload *Sketch, authorization string) (*http.Request, error) {
	var body bytes.Buffer
	err := c.Encoder.Encode(payload, &body, "*/*")
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %s", err)
	}
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	if prod {
		path = prodURL + path
	} else {
		path = devURL + path
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("POST", u.String(), &body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")

	header.Set("Authorization", authorization)

	return req, nil
}

// SearchSketchesPath computes a request path to the search action of sketches.
func SearchSketchesPath() string {
	return fmt.Sprintf("/create/v1/sketches")
}

// SearchSketches Provides a paginated list of sketches filtered according to the params. The page size is 100 items.
func (c *Client) SearchSketches(ctx context.Context, path string, offset *string, owner *string, authorization *string) (*http.Response, error) {
	req, err := c.NewSearchSketchesRequest(ctx, path, offset, owner, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewSearchSketchesRequest create the request corresponding to the search action endpoint of the sketches resource.
func (c *Client) NewSearchSketchesRequest(ctx context.Context, path string, offset *string, owner *string, authorization *string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	if prod {
		path = prodURL + path
	} else {
		path = devURL + path
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	values := u.Query()
	if offset != nil {
		values.Set("offset", *offset)
	}
	if owner != nil {
		values.Set("owner", *owner)
	}
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	header := req.Header
	if authorization != nil {
		header.Set("Authorization", *authorization)
	}
	return req, nil
}

// ShowSketchesPath computes a request path to the show action of sketches.
func ShowSketchesPath(id string) string {
	param0 := id

	return fmt.Sprintf("/create/v1/sketches/%s", param0)
}

// ShowSketches Provides the sketch identified by the :id param.
func (c *Client) ShowSketches(ctx context.Context, path string, authorization *string) (*http.Response, error) {
	req, err := c.NewShowSketchesRequest(ctx, path, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowSketchesRequest create the request corresponding to the show action endpoint of the sketches resource.
func (c *Client) NewShowSketchesRequest(ctx context.Context, path string, authorization *string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	if prod {
		path = prodURL + path
	} else {
		path = devURL + path
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	header := req.Header
	if authorization != nil {
		header.Set("Authorization", *authorization)
	}
	return req, nil
}
