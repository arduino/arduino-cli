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

	jsonData := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]string{
				"content":          base64.StdEncoding.EncodeToString(data),
				"content_encoding": "base64",
			},
			"relationships": map[string]interface{}{
				"resource": map[string]interface{}{
					"data": map[string]string{
						"id":   fmt.Sprintf("o:%s:p:%s:r:%s", organization, project, resource),
						"type": "resources",
					},
				},
			},
			"type": "resource_strings_async_uploads",
		},
	}

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

	var jsonRes map[string]interface{}
	if err = json.Unmarshal(body, &jsonRes); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Started upload of resource file %s\n", resourceFile)
	return jsonRes["data"].(map[string]interface{})["id"].(string)
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

		var jsonRes map[string]interface{}
		if err = json.Unmarshal(body, &jsonRes); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		data := jsonRes["data"].(map[string]interface{})
		attributes := data["attributes"].(map[string]interface{})
		status := attributes["status"].(string)
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
			errs := attributes["errors"].([]map[string]string)
			for _, err := range errs {
				fmt.Printf("%s: %s\n", err["code"], err["detail"])
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
