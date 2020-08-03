package skynet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	gopath "path"
	"path/filepath"
	"strings"

	"gitlab.com/NebulousLabs/errors"
)

type (
	// UploadData contains data to upload, indexed by filenames.
	UploadData map[string]io.Reader

	// UploadOptions contains the options used for uploads.
	UploadOptions struct {
		Options

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

		// SkykeyName is the name of the skykey used to encrypt the upload.
		SkykeyName string
		// SkykeyID is the ID of the skykey used to encrypt the upload.
		SkykeyID string
	}

	// UploadResponse contains the response for uploads.
	UploadResponse struct {
		// Skylink is the returned skylink.
		Skylink string `json:"skylink"`
	}
)

var (
	// DefaultUploadOptions contains the default upload options.
	DefaultUploadOptions = UploadOptions{
		Options: DefaultOptions("/skynet/skyfile"),

		PortalFileFieldName:          "file",
		PortalDirectoryFileFieldName: "files[]",
		CustomFilename:               "",
		CustomDirname:                "",
		SkykeyName:                   "",
		SkykeyID:                     "",
	}
)

// Upload uploads the given generic data.
func Upload(uploadData UploadData, opts UploadOptions) (string, error) {
	// prepare formdata
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	url := makeURL(opts.PortalURL, opts.EndpointPath)

	var fieldname string
	var filename string
	if len(uploadData) == 1 {
		fieldname = opts.PortalFileFieldName
	} else {
		if opts.CustomDirname == "" {
			return "", errors.New("CustomDirname must be set when uploading multiple files")
		}
		fieldname = opts.PortalDirectoryFileFieldName
		filename = opts.CustomDirname
	}
	// The filename is set first and handles including ? in the url string. All
	// subsequent parameters should include an &.
	url = fmt.Sprintf("%s?filename=%s", url, filename)

	// Include the skykey name or id, if given.
	url = fmt.Sprintf("%s&skykeyname=%s", url, opts.SkykeyName)
	url = fmt.Sprintf("%s&skykeyid=%s", url, opts.SkykeyID)

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
		return "", errors.AddContext(err, "could not close writer")
	}
	opts.CustomContentType = writer.FormDataContentType()

	resp, err := executeRequest("POST", url, body, opts.Options)
	if err != nil {
		return "", errors.AddContext(err, "could not execute request")
	}

	respBody, err := parseResponseBody(resp)
	if err != nil {
		return "", errors.AddContext(err, "could not parse response body")
	}

	var apiResponse UploadResponse
	err = json.Unmarshal(respBody.Bytes(), &apiResponse)
	if err != nil {
		return "", errors.AddContext(err, "could not unmarshal response JSON")
	}

	return fmt.Sprintf("%s%s", URISkynetPrefix, apiResponse.Skylink), nil
}

// UploadFile uploads a file to Skynet.
func UploadFile(path string, opts UploadOptions) (skylink string, err error) {
	path = gopath.Clean(path)

	// Open the file.
	file, err := os.Open(gopath.Clean(path)) // Clean again to prevent lint error.
	if err != nil {
		return "", errors.AddContext(err, fmt.Sprintf("could not open file %v", path))
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
