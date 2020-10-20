package skynet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
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

// Upload uploads the given generic data and returns the skylink.
func (sc *SkynetClient) Upload(uploadData UploadData, opts UploadOptions) (skylink string, err error) {
	// prepare formdata
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	var fieldname string
	var filename string
	// Upload as a directory if the dirname is set, even if there is only 1
	// file.
	if len(uploadData) == 1 && opts.CustomDirname == "" {
		fieldname = opts.PortalFileFieldName
	} else {
		if opts.CustomDirname == "" {
			return "", errors.New("CustomDirname must be set when uploading multiple files")
		}
		fieldname = opts.PortalDirectoryFileFieldName
		filename = opts.CustomDirname
	}

	values := url.Values{}
	// Empty values are ignored, but check for "" anyway for clarity.
	if filename != "" {
		// Empty
		values.Set("filename", filename)
	}
	if opts.SkykeyName != "" {
		values.Set("skykeyname", opts.SkykeyName)
	}
	if opts.SkykeyID != "" {
		values.Set("skykeyid", opts.SkykeyID)
	}

	for filename, data := range uploadData {
		// We may need to do a read to determine the Content-Type. Tee the read
		// into a buffer so we can read again.
		var buf bytes.Buffer
		tee := io.TeeReader(data, &buf)
		// Create the form file, inferring the Content-Type.
		part, err := createFormFileContentType(writer, fieldname, filename, tee)
		if err != nil {
			return "", errors.AddContext(err, fmt.Sprintf("could not create form file for file %v", filename))
		}
		// Copy from the buffer and then the rest of the data that hasn't been
		// read.
		_, err = io.Copy(part, &buf)
		_, err2 := io.Copy(part, data)
		if errors.Compose(err, err2) != nil {
			return "", errors.AddContext(err, fmt.Sprintf("could not copy data for file %v", filename))
		}
	}

	err = writer.Close()
	if err != nil {
		return "", errors.AddContext(err, "could not close writer")
	}
	opts.customContentType = writer.FormDataContentType()

	resp, err := sc.executeRequest(
		requestOptions{
			Options: opts.Options,
			method:  "POST",
			reqBody: body,
			query:   values,
		},
	)
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

// UploadFile uploads a file to Skynet and returns the skylink.
func (sc *SkynetClient) UploadFile(path string, opts UploadOptions) (skylink string, err error) {
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

	return sc.Upload(uploadData, opts)
}

// UploadDirectory uploads a local directory to Skynet and returns the skylink.
func (sc *SkynetClient) UploadDirectory(path string, opts UploadOptions) (skylink string, err error) {
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
	basepath := path
	if basepath != "/" {
		basepath += "/"
	}
	for _, filepath := range files {
		file, err := os.Open(gopath.Clean(filepath)) // Clean again to prevent lint error.
		if err != nil {
			return "", errors.AddContext(err, "error opening file")
		}
		// Remove the base path before uploading. Any ending '/' was removed
		// from `path` with `Clean`.
		filepath = strings.TrimPrefix(filepath, basepath)
		uploadData[filepath] = file
	}

	return sc.Upload(uploadData, opts)
}

// createFormFileContentType is based on multipart.Writer.CreateFormFile, except
// it properly sets the content types.
func createFormFileContentType(w *multipart.Writer, fieldname, filename string, file io.Reader) (io.Writer, error) {
	escapeQuotes := func(s string) string {
		var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")
		return quoteEscaper.Replace(s)
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldname), escapeQuotes(filename)))
	contentType, err := getFileContentType(filename, file)
	if err != nil {
		return nil, err
	}
	h.Set("Content-Type", contentType)
	return w.CreatePart(h)
}

// getFileContentType extracts the content type from a given file.
func getFileContentType(filename string, file io.Reader) (string, error) {
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType != "" {
		return contentType, nil
	}

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Always returns a valid content-type by returning
	// "application/octet-stream" if no others seemed to match.
	contentType = http.DetectContentType(buffer)

	return contentType, nil
}
