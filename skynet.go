package skynet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	gopath "path"
	"path/filepath"
	"strings"

	"gitlab.com/NebulousLabs/errors"
)

type (
	// DownloadOptions contains the options used for downloads.
	DownloadOptions struct {
		// PortalDownloadPath is the relative URL path of the download endpoint.
		PortalDownloadPath string
		// PortalURL is the URL of the portal to use.
		PortalURL string
	}

	// UploadData contains data to upload, indexed by filenames.
	UploadData map[string]io.Reader

	// UploadOptions contains the options used for uploads.
	UploadOptions struct {
		// PortalURL is the URL of the portal to use.
		PortalURL string
		// PortalUploadPath is the relative URL path of the upload endpoint.
		PortalUploadPath string
		// PortalFileFieldName is the fieldName for files on the portal.
		PortalFileFieldName string
		// PortalDirectoryFileFieldName is the fieldName for directory files on
		// the portal.
		PortalDirectoryFileFieldName string

		// CustomFilename is the custom filename to use for the upload. If this
		// is empty, the filename of the file being uploaded will be used by
		// default.
		CustomFilename string
		// CustomDirname is the custom name of the directory. If this is empty,
		// the base name of the directory being uploaded will be used by
		// default.
		CustomDirname string
	}

	// UploadResponse contains the response for uploads.
	UploadResponse struct {
		// Skylink is the returned skylink.
		Skylink string `json:"skylink"`
	}
)

const (
	// URISkynetPrefix is the URI prefix for Skynet.
	URISkynetPrefix = "sia://"
)

var (
	// DefaultDownloadOptions contains the default download options.
	DefaultDownloadOptions = DownloadOptions{
		PortalURL:          "https://siasky.net",
		PortalDownloadPath: "/",
	}

	// DefaultUploadOptions contains the default upload options.
	DefaultUploadOptions = UploadOptions{
		PortalURL:                    "https://siasky.net",
		PortalUploadPath:             "/skynet/skyfile",
		PortalFileFieldName:          "file",
		PortalDirectoryFileFieldName: "files[]",
	}
)

// Upload uploads the given generic data.
func Upload(uploadData UploadData, opts UploadOptions) (string, error) {
	// prepare formdata
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	url := fmt.Sprintf("%s/%s", strings.TrimRight(opts.PortalURL, "/"), strings.TrimLeft(opts.PortalUploadPath, "/"))

	var fieldname string
	if len(uploadData) == 1 {
		fieldname = opts.PortalFileFieldName
	} else {
		if opts.CustomDirname == "" {
			return "", errors.New("CustomDirname must be set when uploading multiple files")
		}
		fieldname = opts.PortalDirectoryFileFieldName
		url = fmt.Sprintf("%s?filename=%s", url, opts.CustomDirname)
	}

	for filename, data := range uploadData {
		part, err := writer.CreateFormFile(fieldname, filename)
		if err != nil {
			return "", errors.AddContext(err, fmt.Sprintf("could not create form file for file %v", filename))
		}
		_, err = io.Copy(part, data)
		if err != nil {
			return "", errors.AddContext(err, fmt.Sprintf("could not copy data for file %v", filename))
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
	err = resp.Body.Close()
	if err != nil {
		return "", err
	}

	var apiResponse UploadResponse
	err = json.Unmarshal(body.Bytes(), &apiResponse)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("sia://%s", apiResponse.Skylink), nil
}

// UploadFile uploads a file to Skynet.
func UploadFile(path string, opts UploadOptions) (skylink string, err error) {
	path = gopath.Clean(path)

	// Open the file.
	file, err := os.Open(gopath.Clean(path)) // Clean again to prevent lint error.
	if err != nil {
		return "", err
	}
	defer func() {
		err = errors.Extend(err, file.Close())
	}()

	// Set filename.
	filename := filepath.Base(path)
	if opts.CustomFilename != "" {
		filename = opts.CustomFilename
	}

	uploadData := make(UploadData)
	uploadData[filename] = file

	return Upload(uploadData, opts)
}

// UploadDirectory uploads a local directory to Skynet.
func UploadDirectory(path string, opts UploadOptions) (string, error) {
	path = gopath.Clean(path)

	// Verify the given path is a directory.
	info, err := os.Stat(path)
	if err != nil {
		return "", errors.AddContext(err, "error retrieving path info")
	}
	if !info.IsDir() {
		return "", fmt.Errorf("given path %v is not a directory", path)
	}

	// Find all files in the given directory.
	files, err := walkDirectory(path)
	if err != nil {
		return "", errors.AddContext(err, "error walking directory")
	}

	// Set DirName.
	if opts.CustomDirname == "" {
		opts.CustomDirname = filepath.Base(path)
	}

	// prepare formdata
	uploadData := make(UploadData)
	for _, filepath := range files {
		file, err := os.Open(gopath.Clean(filepath)) // Clean again to prevent lint error.
		if err != nil {
			return "", errors.AddContext(err, "error opening file")
		}
		// Remove the base path before uploading. Any ending '/' was removed
		// from `path` with `Clean`.
		basepath := path
		if basepath != "/" {
			basepath += "/"
		}
		filepath = strings.TrimPrefix(filepath, basepath)
		uploadData[filepath] = file
	}

	return Upload(uploadData, opts)
}

// Download downloads generic data.
func Download(skylink string, opts DownloadOptions) (io.ReadCloser, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s", strings.TrimRight(opts.PortalURL, "/"), strings.TrimPrefix(skylink, "sia://")))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// DownloadFile downloads a file from Skynet to path.
func DownloadFile(path, skylink string, opts DownloadOptions) (err error) {
	path = gopath.Clean(path)

	downloadData, err := Download(skylink, opts)

	if err != nil {
		return
	}
	defer func() {
		err = errors.Extend(err, downloadData.Close())
	}()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Extend(err, out.Close())
	}()

	_, err = io.Copy(out, downloadData)
	return err
}

// walkDirectory walks a given directory recursively, returning the paths of all
// files found.
func walkDirectory(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if subpath == path {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, subpath)
		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return files, nil
}
