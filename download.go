package skynet

import (
	"bytes"
	"io"
	"net/url"
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
func (sc *SkynetClient) Download(skylink string, opts DownloadOptions) (io.ReadCloser, error) {
	skylink = strings.TrimPrefix(skylink, URISkynetPrefix)

	values := url.Values{}
	values.Set("skykeyname", opts.SkykeyName)
	values.Set("skykeyid", opts.SkykeyID)

	resp, err := sc.executeRequest(
		requestOptions{
			Options:   opts.Options,
			method:    "GET",
			reqBody:   &bytes.Buffer{},
			extraPath: skylink,
			query:     values,
		},
	)
	if err != nil {
		return nil, errors.AddContext(err, "could not execute request")
	}

	return resp.Body, nil
}

// DownloadFile downloads a file from Skynet to path.
func (sc *SkynetClient) DownloadFile(path, skylink string, opts DownloadOptions) (err error) {
	path = gopath.Clean(path)

	downloadData, err := sc.Download(skylink, opts)
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
func (sc *SkynetClient) Metadata(skylink string, opts MetadataOptions) error {
	panic("Not implemented")
}
