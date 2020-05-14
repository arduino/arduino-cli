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
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

var pushTransifexCommand = &cobra.Command{
	Use:   "push [catalog folder]",
	Short: "pushes the translation files to transifex",
	Args:  cobra.ExactArgs(1),
	Run:   pushCatalog,
}

func pushFile(folder, lang, url string) {
	filename := path.Join(folder, lang+".po")
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(lang, filepath.Base(filename))

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	_, err = io.Copy(part, file)
	writer.WriteField("file_type", "po")

	req, err := http.NewRequest("PUT", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	req.SetBasicAuth("api", apiKey)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	resp.Body.Close()
}

func pushCatalog(cmd *cobra.Command, args []string) {
	folder := args[0]

	pushFile(
		folder,
		"en",
		fmt.Sprintf(
			"https://www.transifex.com/api/2/project/%s/resource/%s/content/",
			project,
			resource,
		),
	)
}
