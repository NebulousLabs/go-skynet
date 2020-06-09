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
	// UploadOptions contains the options used for uploads.
	UploadOptions struct {
		// PortalURL is the URL of the portal to use.
		PortalURL string
		// PortalUploadPath is the path on the portal of the upload endpoint.
		PortalUploadPath string
		// PortalFileFieldname is the fieldname for files on the portal.
		PortalFileFieldname string
		// PortalDirectoryFileFieldname is the fieldname for directory files on
		// the portal.
		PortalDirectoryFileFieldname string
		// CustomFilename is the custom filename to use for the upload. If this
		// is empty, the filename of the file being uploaded will be used by
		// default.
		CustomFilename string
	}

	// DownloadOptions contains the options used for downloads.
	DownloadOptions struct {
		// PortalURL is the URL of the portal to use.
		PortalURL string
	}

	// uploadResponse contains the response for uploads.
	uploadResponse struct {
		// Skylink is the returned skylink.
		Skylink string `json:"skylink"`
	}
)

var (
	// DefaultUploadOptions contains the default upload options.
	DefaultUploadOptions = UploadOptions{
		PortalURL:                    "https://siasky.net",
		PortalUploadPath:             "/skynet/skyfile",
		PortalFileFieldname:          "file",
		PortalDirectoryFileFieldname: "files[]",
		CustomFilename:               "",
	}

	// DefaultDownloadOptions contains the default download options.
	DefaultDownloadOptions = DownloadOptions{
		PortalURL: "https://siasky.net",
	}
)

// UploadFile uploads a file to Skynet.
func UploadFile(path string, opts UploadOptions) (skylink string, err error) {
	path = gopath.Clean(path)

	// open the file
	file, err := os.Open(gopath.Clean(path))
	if err != nil {
		return "", err
	}
	defer func() {
		err = errors.Extend(err, file.Close())
	}()

	// set filename
	var filename string
	if opts.CustomFilename != "" {
		filename = opts.CustomFilename
	} else {
		filename = filepath.Base(path)
	}

	// prepare formdata
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(opts.PortalFileFieldname, filename)
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

	url := fmt.Sprintf("%s/%s", strings.TrimRight(opts.PortalURL, "/"), strings.TrimLeft(opts.PortalUploadPath, "/"))

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

	var apiResponse uploadResponse
	err = json.Unmarshal(body.Bytes(), &apiResponse)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("sia://%s", apiResponse.Skylink), nil
}

// UploadDirectory uploads a directory to Skynet.
func UploadDirectory(path string, opts UploadOptions) (string, error) {
	path = gopath.Clean(path)

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
	if opts.CustomFilename != "" {
		filename = opts.CustomFilename
	} else {
		filename = filepath.Base(path)
	}

	// prepare formdata
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, filepath := range files {
		file, err := os.Open(gopath.Clean(filepath))
		if err != nil {
			return "", err
		}
		part, err := writer.CreateFormFile(opts.PortalDirectoryFileFieldname, filepath)
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

	url := fmt.Sprintf("%s/%s?filename=%s", strings.TrimRight(opts.PortalURL, "/"), strings.TrimLeft(opts.PortalUploadPath, "/"), filename)

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

	var apiResponse uploadResponse
	err = json.Unmarshal(body.Bytes(), &apiResponse)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("sia://%s", apiResponse.Skylink), nil
}

// DownloadFile downloads a file from Skynet.
func DownloadFile(path, skylink string, opts DownloadOptions) (err error) {
	path = gopath.Clean(path)

	resp, err := http.Get(fmt.Sprintf("%s/%s", strings.TrimRight(opts.PortalURL, "/"), strings.TrimPrefix(skylink, "sia://")))
	if err != nil {
		return
	}
	defer func() {
		err = errors.Extend(err, resp.Body.Close())
	}()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Extend(err, out.Close())
	}()

	_, err = io.Copy(out, resp.Body)
	return
}

// walkDirectory walks a given directory recursively, returning the paths of all
// files found.
func walkDirectory(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fullpath := filepath.Join(path, subpath)
		if info.IsDir() {
			subfiles, err := walkDirectory(fullpath)
			if err != nil {
				return err
			}
			files = append(files, subfiles...)
			return nil
		}
		files = append(files, fullpath)
		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return files, nil
}
