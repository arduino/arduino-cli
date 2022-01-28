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

package transifex

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

var pushTransifexCommand = &cobra.Command{
	Use:   "push [catalog folder]",
	Short: "pushes the translation files to transifex",
	Args:  cobra.ExactArgs(1),
	Run:   pushCatalog,
}

// uploadSourceFile starts an async upload of resourceFile.
// Returns an id to monitor the upload status.
func uploadSourceFile(resourceFile *paths.Path) string {
	url := mainEndpoint + "resource_strings_async_uploads"
	data, err := resourceFile.ReadFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	type jsonReq struct {
		Data struct {
			Attributes struct {
				Content         string `json:"content"`
				ContentEncoding string `json:"content_encoding"`
			} `json:"attributes"`
			Relationships struct {
				Resource struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
				} `json:"resource"`
			} `json:"relationships"`
			Type string `json:"type"`
		} `json:"data"`
	}
	jsonData := jsonReq{}
	jsonData.Data.Type = "resource_strings_async_uploads"
	jsonData.Data.Attributes.Content = base64.StdEncoding.EncodeToString(data)
	jsonData.Data.Attributes.ContentEncoding = "base64"
	jsonData.Data.Relationships.Resource.Data.ID = fmt.Sprintf("o:%s:p:%s:r:%s", organization, project, resource)
	jsonData.Data.Relationships.Resource.Data.Type = "resources"

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	addHeaders(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var jsonRes struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err = json.Unmarshal(body, &jsonRes); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Started upload of resource file %s\n", resourceFile)
	return jsonRes.Data.ID
}

func checkUploadStatus(uploadID string) {
	url := mainEndpoint + "resource_strings_async_uploads/" + uploadID
	// The upload request status must be asked from time to time, if it's
	// still pending we try again using exponentional backoff starting from 2.5 seconds.
	backoff := 2500 * time.Millisecond

	for {
		req, err := http.NewRequest(
			"GET",
			url,
			nil,
		)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		addHeaders(req)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var body []byte
		{
			defer res.Body.Close()
			body, err = io.ReadAll(res.Body)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		var jsonRes struct {
			Data struct {
				Attributes struct {
					Status string `json:"status"`
					Errors []struct {
						Code   string `json:"code"`
						Detail string `json:"detail"`
					} `json:"errors"`
				} `json:"attributes"`
			} `json:"data"`
		}
		if err = json.Unmarshal(body, &jsonRes); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		status := jsonRes.Data.Attributes.Status
		switch status {
		case "succeeded":
			fmt.Println("Resource file uploaded")
			return
		case "pending":
			fallthrough
		case "processing":
			fmt.Printf("Current status: %s\n", status)
			time.Sleep(backoff)
			backoff = backoff * 2
			// Request the status again
			continue
		case "failed":
			for _, err := range jsonRes.Data.Attributes.Errors {
				fmt.Printf("%s: %s\n", err.Code, err.Detail)
			}
			os.Exit(1)
		}
		fmt.Println("Status request failed in an unforeseen way")
		os.Exit(1)
	}
}

func pushFile(resourceFile *paths.Path) {
	uploadID := uploadSourceFile(resourceFile)
	checkUploadStatus(uploadID)
}

func pushCatalog(cmd *cobra.Command, args []string) {
	folder := args[0]

	pushFile(paths.New(folder, "en.po"))
}
