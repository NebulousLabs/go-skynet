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
	// Skykey contains information about a skykey.
	Skykey struct {
		Skykey string `json:"skykey"`
		Name   string `json:"name"`
		ID     string `json:"id"`
		Type   string `json:"type"`
	}

	// UploadData contains data to upload, indexed by filenames.
	UploadData map[string]io.Reader

	// Options structs.

	// ConnectionOptions contains options used for connecting to Skynet.
	ConnectionOptions struct {
		// PortalURL is the URL of the portal to use.
		PortalURL string
		// CustomUserAgent is the custom user agent to use.
		CustomUserAgent string
	}

	// AddSkykeyOptions contains the options used for addskykey.
	AddSkykeyOptions struct {
		ConnectionOptions
		// PortalAddSkykeyPath is the relative URL path of the addskykey
		// endpoint.
		PortalAddSkykeyPath string
	}
	// CreateSkykeyOptions contains the options used for createskykey.
	CreateSkykeyOptions struct {
		ConnectionOptions
		// PortalCreateSkykeyPath is the relative URL path of the createskykey
		// endpoint.
		PortalCreateSkykeyPath string
	}
	// GetSkykeyOptions contains the options used for skykey GET.
	GetSkykeyOptions struct {
		ConnectionOptions
		// PortalGetSkykeyPath is the relative URL path of the skykey GET
		// endpoint.
		PortalGetSkykeyPath string
	}
	// ListSkykeysOptions contains the options used for skykeys GET.
	ListSkykeysOptions struct {
		ConnectionOptions
		// PortalListSkykeysPath is the relative URL path of the skykeys
		// endpoint.
		PortalListSkykeysPath string
	}

	// DownloadOptions contains the options used for downloads.
	DownloadOptions struct {
		ConnectionOptions
		// PortalDownloadPath is the relative URL path of the download endpoint.
		PortalDownloadPath string
	}

	// UploadOptions contains the options used for uploads.
	UploadOptions struct {
		ConnectionOptions

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

		// SkykeyName is the name of the skykey used to encrypt the upload.
		SkykeyName string
		// SkykeyID is the ID of the skykey used to encrypt the upload.
		SkykeyID string
	}

	// Response structs

	// ErrorResponse contains the response for an error.
	ErrorResponse struct {
		// Message is the error message of the response.
		Message string `json:"message"`
	}

	// CreateSkykeyResponse contains the response for creating a skykey.
	CreateSkykeyResponse Skykey
	// GetSkykeyResponse contains the response for getting a skykey.
	GetSkykeyResponse Skykey
	// ListSkykeysResponse contains the response for listing skykeys.
	ListSkykeysResponse struct {
		// Skykeys is the returned list of skykeys.
		Skykeys []Skykey `json:"skykeys"`
	}

	// UploadResponse contains the response for uploads.
	UploadResponse struct {
		// Skylink is the returned skylink.
		Skylink string `json:"skylink"`
	}
)

const (
	// DefaultPortalURL is the default URL of the portal to use.
	DefaultPortalURL = "https://siasky.net"

	// URISkynetPrefix is the URI prefix for Skynet.
	URISkynetPrefix = "sia://"
)

var (
	// DefaultConnectionOptions contains the default connection options.
	DefaultConnectionOptions = ConnectionOptions{
		PortalURL: DefaultPortalURL,
	}

	// DefaultAddSkykeyOptions contains the default addskykey options.
	DefaultAddSkykeyOptions = AddSkykeyOptions{
		ConnectionOptions:   DefaultConnectionOptions,
		PortalAddSkykeyPath: "/skynet/addskykey",
	}
	// DefaultCreateSkykeyOptions contains the default createskykey options.
	DefaultCreateSkykeyOptions = CreateSkykeyOptions{
		ConnectionOptions:      DefaultConnectionOptions,
		PortalCreateSkykeyPath: "/skynet/createskykey",
	}
	// DefaultGetSkykeyOptions contains the default skykey GET options.
	DefaultGetSkykeyOptions = GetSkykeyOptions{
		ConnectionOptions:   DefaultConnectionOptions,
		PortalGetSkykeyPath: "/skynet/skykey",
	}
	// DefaultListSkykeysOptions contains the default skykeys options.
	DefaultListSkykeysOptions = ListSkykeysOptions{
		ConnectionOptions:     DefaultConnectionOptions,
		PortalListSkykeysPath: "/skynet/skykeys",
	}

	// DefaultDownloadOptions contains the default download options.
	DefaultDownloadOptions = DownloadOptions{
		ConnectionOptions:  DefaultConnectionOptions,
		PortalDownloadPath: "/",
	}

	// DefaultUploadOptions contains the default upload options.
	DefaultUploadOptions = UploadOptions{
		ConnectionOptions:            DefaultConnectionOptions,
		PortalUploadPath:             "/skynet/skyfile",
		PortalFileFieldName:          "file",
		PortalDirectoryFileFieldName: "files[]",
	}

	// ErrResponseError is the error for a response with a status code >= 400.
	ErrResponseError = errors.New("error response")
)

// AddSkykey stores the given base-64 encoded skykey with the renter's skykey
// manager.
func AddSkykey(skykey string, opts AddSkykeyOptions) error {
	body := &bytes.Buffer{}
	url := makeURL(opts.PortalURL, opts.PortalAddSkykeyPath)
	url = fmt.Sprintf("%s?skykey=%s", url, skykey)

	req, err := makeRequest(opts.ConnectionOptions, "POST", url, body)
	if err != nil {
		return errors.AddContext(err, "could not make request")
	}

	// Add the skykey.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.AddContext(err, "could not execute POST")
	}
	if resp.StatusCode >= 400 {
		return errors.AddContext(makeResponseError(resp), "error code received")
	}

	return nil
}

// CreateSkykey returns a new skykey created and stored under the given name
// with the given type. skykeyType can be either "public-id" or "private-id".
func CreateSkykey(name, skykeyType string, opts CreateSkykeyOptions) (Skykey, error) {
	body := &bytes.Buffer{}
	url := makeURL(opts.PortalURL, opts.PortalCreateSkykeyPath)
	url = fmt.Sprintf("%s?name=%s&type=%s", url, name, skykeyType)

	req, err := makeRequest(opts.ConnectionOptions, "POST", url, body)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not make request")
	}

	// Create the skykey.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not execute POST")
	}
	if resp.StatusCode >= 400 {
		return Skykey{}, errors.AddContext(makeResponseError(resp), "error code received")
	}

	// parse the response
	body = &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not parse response body")
	}
	err = resp.Body.Close()
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not close response body")
	}

	var apiResponse CreateSkykeyResponse
	err = json.Unmarshal(body.Bytes(), &apiResponse)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not unmarshal response JSON")
	}

	return Skykey(apiResponse), nil
}

// GetSkykey returns the given skykey. One of either name or id must be provided
// -- the one that is not provided should be "".
func GetSkykey(name, id string, opts GetSkykeyOptions) (Skykey, error) {
	body := &bytes.Buffer{}
	url := makeURL(opts.PortalURL, opts.PortalGetSkykeyPath)
	url = fmt.Sprintf("%s?name=%s&id=%s", url, name, id)

	req, err := makeRequest(opts.ConnectionOptions, "GET", url, body)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not make request")
	}

	// Create the skykey.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not execute GET")
	}
	if resp.StatusCode >= 400 {
		return Skykey{}, errors.AddContext(makeResponseError(resp), "error code received")
	}

	// parse the response
	body = &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not parse response body")
	}
	err = resp.Body.Close()
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not close response body")
	}

	var apiResponse GetSkykeyResponse
	err = json.Unmarshal(body.Bytes(), &apiResponse)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not unmarshal response JSON")
	}

	return Skykey(apiResponse), nil
}

// ListSkykeys returns a list of all skykeys.
func ListSkykeys(opts ListSkykeysOptions) ([]Skykey, error) {
	url := makeURL(opts.PortalURL, opts.PortalListSkykeysPath)

	req, err := makeRequest(opts.ConnectionOptions, "GET", url, &bytes.Buffer{})
	if err != nil {
		return nil, errors.AddContext(err, "could not make request")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.AddContext(err, "could not execute GET")
	}
	if resp.StatusCode >= 400 {
		return nil, errors.AddContext(makeResponseError(resp), "error code received")
	}

	// parse the response
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return nil, errors.AddContext(err, "could not read from response body")
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, errors.AddContext(err, "could not close response body")
	}

	var apiResponse ListSkykeysResponse
	err = json.Unmarshal(body.Bytes(), &apiResponse)
	if err != nil {
		return nil, errors.AddContext(err, "could not unmarshal response JSON")
	}

	return apiResponse.Skykeys, nil
}

// Upload uploads the given generic data.
func Upload(uploadData UploadData, opts UploadOptions) (string, error) {
	// prepare formdata
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	url := makeURL(opts.PortalURL, opts.PortalUploadPath)

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
	// Always send the filename even if it's empty. This lets us always pass
	// more URL parameters using &.
	url = fmt.Sprintf("%s?filename=%s", url, filename)

	// Include the skykey name or id, if given.
	if opts.SkykeyName != "" {
		url = fmt.Sprintf("%s&skykeyname=%s", url, opts.SkykeyName)
	}
	if opts.SkykeyID != "" {
		url = fmt.Sprintf("%s&skykeyid=%s", url, opts.SkykeyID)
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
		return "", errors.AddContext(err, "could not close writer")
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", errors.AddContext(err, "could not create POST request")
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// upload the file to skynet
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.AddContext(err, "could not execute POST")
	}
	if resp.StatusCode >= 400 {
		return "", errors.AddContext(makeResponseError(resp), "error code received")
	}

	// parse the response
	body = &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return "", errors.AddContext(err, "could not parse response body")
	}
	err = resp.Body.Close()
	if err != nil {
		return "", errors.AddContext(err, "could not close response body")
	}

	var apiResponse UploadResponse
	err = json.Unmarshal(body.Bytes(), &apiResponse)
	if err != nil {
		return "", errors.AddContext(err, "could not unmarshal response JSON")
	}

	return fmt.Sprintf("sia://%s", apiResponse.Skylink), nil
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

// Download downloads generic data.
func Download(skylink string, opts DownloadOptions) (io.ReadCloser, error) {
	url := makeURL(opts.PortalURL, opts.PortalDownloadPath)
	url = fmt.Sprintf("%s/%s", strings.TrimRight(url, "/"), strings.TrimPrefix(skylink, "sia://"))

	req, err := makeRequest(opts.ConnectionOptions, "GET", url, &bytes.Buffer{})
	if err != nil {
		return nil, errors.AddContext(err, "could not make request")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.AddContext(err, "could not execute GET")
	}
	if resp.StatusCode >= 400 {
		return nil, errors.AddContext(makeResponseError(resp), "error code received")
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

// makeResponseError makes an error given an error response.
func makeResponseError(resp *http.Response) error {
	body := &bytes.Buffer{}
	_, err := body.ReadFrom(resp.Body)
	if err != nil {
		return errors.AddContext(err, "could not read from response body")
	}
	err = resp.Body.Close()
	if err != nil {
		return errors.AddContext(err, "could not close response body")
	}

	var apiResponse ErrorResponse
	message := string(body.Bytes())
	err = json.Unmarshal(body.Bytes(), &apiResponse)
	if err == nil {
		message = apiResponse.Message
	}

	context := fmt.Sprintf("%v response from %v: %v", resp.StatusCode, resp.Request.Method, message)
	return errors.AddContext(ErrResponseError, context)
}

// makeRequest makes a request given the ConnectionOptions.
func makeRequest(copts ConnectionOptions, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.AddContext(err, fmt.Sprintf("could not create %v request", method))
	}
	if copts.CustomUserAgent != "" {
		req.Header.Set("User-Agent", copts.CustomUserAgent)
	}
	return req, nil
}

// makeURL makes a URL from the given parts.
func makeURL(portalURL, portalPath string) string {
	return fmt.Sprintf("%s/%s", strings.TrimRight(portalURL, "/"), strings.TrimLeft(portalPath, "/"))
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
