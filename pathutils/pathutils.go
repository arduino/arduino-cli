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
 * Copyright 2018 ARDUINO AG (http://www.arduino.cc/)
 */

package pathutils

import (
	"fmt"
	"os"
	"path/filepath"
)

// Path is an object that determine a path
type Path interface {
	// Get tries to determine the path or return an error if fails
	Get() (string, error)
}

// NewConstPath creates a constant Path that always returns the specified
// path, errors will always be nil
func NewConstPath(label string, path string, createIfMissing bool) Path {
	return &basicPath{
		Label: label,
		ProviderFunction: func() (string, error) {
			return path, nil
		},
		CreateIfMissing: createIfMissing,
	}
}

// NewPath return a Path object that use the providerFunction to determine
// the current path. If CreateIfMissing is true a new folder will be created
// if the returned path doesn't exists
func NewPath(label string, providerFunction func() (string, error), createIfMissing bool) Path {
	return &basicPath{
		Label:            label,
		ProviderFunction: providerFunction,
		CreateIfMissing:  createIfMissing,
	}
}

// NewSubPath use Path as root and appends subPath to determine the final path.
// If CreateIfMissing is true and the resulting path+subPath doesn't exists a
// new folder will be created if.
func NewSubPath(label string, path Path, subPath string, createIfMissing bool) Path {
	return &basicSubPath{
		Label:           label,
		Path:            path,
		SubPath:         subPath,
		CreateIfMissing: createIfMissing,
	}
}

type basicPath struct {
	Label            string
	ProviderFunction func() (string, error)
	CreateIfMissing  bool

	cachedPath string
}

func (p *basicPath) Get() (string, error) {
	if p.cachedPath != "" {
		return p.cachedPath, nil
	}
	path, err := p.ProviderFunction()
	if err != nil {
		return "", fmt.Errorf("Cannot get %s folder: %s", p.Label, err)
	}
	if p.CreateIfMissing {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				return "", fmt.Errorf("Creating %s folder: %s", path, err)
			}
		} else if err != nil {
			return "", fmt.Errorf("Checking if %s exists: %s", path, err)
		}
	}
	p.cachedPath = path
	return p.cachedPath, nil
}

type basicSubPath struct {
	Label           string
	Path            Path
	SubPath         string
	CreateIfMissing bool

	cachedPath string
}

func (p *basicSubPath) Get() (string, error) {
	if p.cachedPath != "" {
		return p.cachedPath, nil
	}
	path, err := p.Path.Get()
	if err != nil {
		return "", fmt.Errorf("Cannot get %s folder: %s", p.Label, err)
	}
	path = filepath.Join(path, p.SubPath)
	if p.CreateIfMissing {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				return "", fmt.Errorf("Creating %s folder: %s", path, err)
			}
		} else if err != nil {
			return "", fmt.Errorf("Checking if %s exists: %s", path, err)
		}
	}
	p.cachedPath = path
	return p.cachedPath, nil
}
