package skynet

import (
	"bytes"
	"encoding/json"
	"errors"
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
		dirname                      string
	}

	FileUploadOptions struct {
		UploadOptions
		customFilename string
	}

	// keys are filenames
	UploadData map[string]io.Reader

	DownloadOptions struct {
		portalUrl string
	}
)

var (
	DefaultUploadOptions = FileUploadOptions{
		UploadOptions: UploadOptions{
			portalUrl:                    "https://siasky.net",
			portalUploadPath:             "/skynet/skyfile",
			portalFileFieldname:          "file",
			portalDirectoryFileFieldname: "files[]",
		},
		customFilename: "",
	}

	DefaultDownloadOptions = DownloadOptions{
		portalUrl: "https://siasky.net",
	}
)

func Upload(uploadData UploadData, opts UploadOptions) (string, error) {
	// prepare formdata
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	url := fmt.Sprintf("%s/%s", strings.TrimRight(opts.portalUrl, "/"), strings.TrimLeft(opts.portalUploadPath, "/"))

	var fieldname string
	if len(uploadData) == 1 {
		fieldname = opts.portalFileFieldname
	} else {
		if opts.dirname == "" {
			return "", errors.New("dirname must be set when uploading multiple files")
		}
		fieldname = opts.portalDirectoryFileFieldname
		url = fmt.Sprintf("%s?filename=%s", url, opts.dirname)
	}

	for filename, data := range uploadData {
		part, err := writer.CreateFormFile(fieldname, filename)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(part, data)
		if err != nil {
			return "", err
		}
	}

	err := writer.Close()
	if err != nil {
		return "", err
	}

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

func UploadFile(path string, opts FileUploadOptions) (string, error) {
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

	uploadData := make(UploadData)
	uploadData[filename] = file

	return Upload(uploadData, opts.UploadOptions)
}

func UploadDirectory(path string, opts FileUploadOptions) (string, error) {
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

	// set dirname
	if opts.customFilename != "" {
		opts.dirname = opts.customFilename
	} else {
		opts.dirname = filepath.Base(path)
	}

	// prepare formdata
	uploadData := make(UploadData)
	for _, fp := range files {
		file, err := os.Open(fp)
		if err != nil {
			return "", err
		}
		uploadData[fp] = file
	}

	return Upload(uploadData, opts.UploadOptions)
}

func Download(skylink string, opts DownloadOptions) (io.ReadCloser, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s", strings.TrimRight(opts.portalUrl, "/"), strings.TrimPrefix(skylink, "sia://")))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func DownloadFile(path, skylink string, opts DownloadOptions) error {
	downloadData, err := Download(skylink, opts)
	if err != nil {
		return err
	}
	defer downloadData.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, downloadData)
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
