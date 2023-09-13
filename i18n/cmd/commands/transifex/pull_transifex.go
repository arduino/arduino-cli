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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
	"github.com/arduino/arduino-cli/i18n"
)

var tr = i18n.Tr

var pullTransifexCommand = &cobra.Command{
	Use:   "pull [catalog folder]",
	Short: "pulls the translation files from transifex",
	Run:   pullCatalog,
}

func getLanguages() []string {
	url := mainEndpoint + fmt.Sprintf("projects/o:%s:p:%s/languages", organization, project)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	addHeaders(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var jsonRes struct {
		Data []struct {
			Attributes struct {
				Code string `json:"code"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &jsonRes); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var languages []string
	for _, object := range jsonRes.Data {
		languages = append(languages, object.Attributes.Code)
	}
	return languages
}

// startTranslationDownload notifies Transifex that we want to start downloading
// the resources file for the specified languageCode.
// Returns an id to monitor the download status.
func startTranslationDownload(languageCode string) string {
	url := mainEndpoint + "resource_translations_async_downloads"

	type jsonReq struct {
		Data struct {
			Relationships struct {
				Language struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
				} `json:"language"`
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
	jsonData.Data.Type = "resource_translations_async_downloads"
	jsonData.Data.Relationships.Language.Data.ID = fmt.Sprintf("l:%s", languageCode)
	jsonData.Data.Relationships.Language.Data.Type = "languages"
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
	return jsonRes.Data.ID
}

// getDownloadURL checks for the download status of the languageCode file specified
// by downloadID.
// It return a URL to download the file when ready.
func getDownloadURL(languageCode, downloadID string) string {
	url := mainEndpoint + "resource_translations_async_downloads/" + downloadID
	// The download request status must be asked from time to time, if it's
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

		client := http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// We handle redirection manually
				return http.ErrUseLastResponse
			},
		}
		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if res.StatusCode == 303 {
			// Return the URL to download translation file
			return res.Header.Get("location")
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		res.Body.Close()

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
			return ""
		case "pending":
			fallthrough
		case "processing":
			fmt.Printf(tr("Current status for language %s: %s\n"), languageCode, status)
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
		fmt.Printf(tr("Status request for language %s failed in an unforeseen way\n"), languageCode)
		os.Exit(1)
	}
}

// download file from url and saves it in folder with the specified fileName
func download(folder, fileName, url string) {
	fmt.Printf(tr("Starting download of %s\n"), fileName)
	filePath := paths.New(folder, fileName)

	res, err := http.DefaultClient.Get(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	filePath.WriteFile(data)
	fmt.Printf(tr("Finished download of %s\n"), fileName)
}

func pullCatalog(cmd *cobra.Command, args []string) {
	languages := getLanguages()
	fmt.Println(tr("translations found:"), languages)

	folder := args[0]

	var wg sync.WaitGroup
	for _, lang := range languages {
		wg.Add(1)
		go func(lang string) {
			downloadID := startTranslationDownload(lang)
			url := getDownloadURL(lang, downloadID)
			download(folder, lang+".po", url)
			wg.Done()
		}(lang)
	}
	wg.Wait()
	fmt.Println(tr("Translation files downloaded"))
}
