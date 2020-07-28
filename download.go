package skynet

import (
	"bytes"
	"fmt"
	"io"
	"os"
	gopath "path"
	"strings"

	"gitlab.com/NebulousLabs/errors"
)

type (
	// DownloadOptions contains the options used for downloads.
	DownloadOptions struct {
		Options

		// SkykeyName is the name of the skykey used to encrypt the upload.
		SkykeyName string
		// SkykeyID is the ID of the skykey used to encrypt the upload.
		SkykeyID string
	}

	// MetadataOptions contains the options used for getting metadata.
	MetadataOptions struct {
		Options
	}
)

var (
	// DefaultDownloadOptions contains the default download options.
	DefaultDownloadOptions = DownloadOptions{
		Options: DefaultOptions("/"),

		SkykeyName: "",
		SkykeyID:   "",
	}

	// DefaultMetadataOptions contains the default getting metadata options.
	DefaultMetadataOptions = MetadataOptions{
		Options: DefaultOptions("/"),
	}
)

// Download downloads generic data.
func Download(skylink string, opts DownloadOptions) (io.ReadCloser, error) {
	url := makeURL(opts.PortalURL, opts.EndpointPath)
	url = fmt.Sprintf("%s/%s", strings.TrimRight(url, "/"), strings.TrimPrefix(skylink, "sia://"))

	// Include the skykey name or id, if given.
	url = fmt.Sprintf("%s?skykeyname=%s", url, opts.SkykeyName)
	url = fmt.Sprintf("%s&skykeyid=%s", url, opts.SkykeyID)

	resp, err := executeRequest(opts.Options, "GET", url, &bytes.Buffer{})
	if err != nil {
		return nil, errors.AddContext(err, "could not execute request")
	}

	return resp.Body, nil
}

// DownloadFile downloads a file from Skynet to path.
func DownloadFile(path, skylink string, opts DownloadOptions) (err error) {
	path = gopath.Clean(path)

	downloadData, err := Download(skylink, opts)
	if err != nil {
		return errors.AddContext(err, "could not download data")
	}
	defer func() {
		err = errors.Extend(err, downloadData.Close())
	}()

	out, err := os.Create(path)
	if err != nil {
		return errors.AddContext(err, "could not create file at "+path)
	}
	defer func() {
		err = errors.Extend(err, out.Close())
	}()

	_, err = io.Copy(out, downloadData)
	return errors.AddContext(err, "could not copy data to file at "+path)
}

// Metadata downloads metadata from the given skylink.
func Metadata(skylink string, opts MetadataOptions) error {
	panic("Not implemented")
}
