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
		portalUrl                    string
		portalUploadPath             string
		portalFileFieldname          string
		portalDirectoryFileFieldname string
		customFilename               string
	}

	DownloadOptions struct {
		portalUrl string
	}
)

var (
	DefaultUploadOptions = UploadOptions{
		portalUrl:                    "https://siasky.net",
		portalUploadPath:             "/skynet/skyfile",
		portalFileFieldname:          "file",
		portalDirectoryFileFieldname: "files[]",
		customFilename:               "",
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

func UploadDirectory(path string, opts UploadOptions) (string, error) {
	// verify the given path is a directory
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("Given path %v is not a directory", path)
	}

	// find all files in the given directory
	files, err := walkDirectory(path)
	if err != nil {
		return "", err
	}

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
	for _, filepath := range files {
		file, err := os.Open(filepath)
		if err != nil {
			return "", err
		}
		part, err := writer.CreateFormFile(opts.portalDirectoryFileFieldname, filepath)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return "", err
		}
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/%s?filename=%s", strings.TrimRight(opts.portalUrl, "/"), strings.TrimLeft(opts.portalUploadPath, "/"), filename)

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
	resp, err := http.Get(fmt.Sprintf("%s/%s", strings.TrimRight(opts.portalUrl, "/"), strings.TrimPrefix(skylink, "sia://")))
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

func walkDirectory(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return files, nil
}
