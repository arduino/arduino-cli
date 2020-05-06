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
	Use:   "push -l pt_BR [catalog folder]",
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

	for _, lang := range languages {
		pushFile(
			folder,
			lang,
			fmt.Sprintf(
				"https://www.transifex.com/api/2/project/%s/resource/%s/translation/%s/",
				project, resource, lang,
			),
		)
	}
}
