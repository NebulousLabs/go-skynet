package skynet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type (
	UploadReponse struct {
		Skylink string `json:"skylink"`
	}

	UploadOptions struct {
		portalUrl           string
		portalUploadPath    string
		portalFileFieldname string
		customFilename      string
	}

	DownloadOptions struct {
		portalUrl string
	}
)

var (
	DefaultUploadOptions = UploadOptions{
		portalUrl:           "https://siasky.net",
		portalUploadPath:    "/api/skyfile",
		portalFileFieldname: "file",
		customFilename:      "",
	}

	DefaultDownloadOptions = DownloadOptions{
		portalUrl: "https://siasky.net",
	}
)

func UploadFile(path string, opts UploadOptions) (string, error) {
	// open the file
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// set filename
	var filename string
	if opts.customFilename != "" {
		filename = opts.customFilename
	} else {
		filename = filepath.Base(path)
	}

	// prepare formdata
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(opts.portalFileFieldname, filename)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}

	// prepare the request
	url := fmt.Sprintf("%s/%s", strings.TrimRight(opts.portalUrl, "/"), strings.TrimLeft(opts.portalUploadPath, "/"))
	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		return "", err
	}

	// upload the file to skynet
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// parse the response
	body = &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	var apiResponse UploadReponse
	err = json.Unmarshal(body.Bytes(), &apiResponse)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("sia://%s", apiResponse.Skylink), nil
}

func DownloadFile(path, skylink string, opts DownloadOptions) error {
	resp, err := http.Get(fmt.Sprintf("%s/%s", strings.TrimRight(opts.portalUrl, "/"), strings.trimPrefix(skylink, "sia://")))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
