package skynet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/NebulousLabs/errors"
)

type (
	// ErrorResponse contains the response for an error.
	ErrorResponse struct {
		// Message is the error message of the response.
		Message string `json:"message"`
	}

	// Options contains options used for connecting to a Skynet portal and
	// endpoint.
	Options struct {
		// PortalURL is the URL of the portal to use.
		PortalURL string
		// EndpointPath is the relative URL path of the portal endpoint to
		// contact.
		EndpointPath string

		// APIKey is the API password to use for authentication.
		APIKey string
		// CustomUserAgent is the custom user agent to use.
		CustomUserAgent string

		// customContentType is the custom content type to use. Set internally.
		customContentType string
	}
)

const (
	// DefaultPortalURL is the default URL of the portal to use.
	DefaultPortalURL = "https://siasky.net"

	// URISkynetPrefix is the URI prefix for Skynet.
	URISkynetPrefix = "sia://"
)

var (
	// ErrResponseError is the error for a response with a status code >= 400.
	ErrResponseError = errors.New("error response")
)

// DefaultOptions returns the default options with the given endpoint path.
func DefaultOptions(endpointPath string) Options {
	return Options{
		PortalURL:    DefaultPortalURL,
		EndpointPath: endpointPath,
	}
}

// executeRequest makes and executes a request given the Options.
func executeRequest(method, url string, reqBody io.Reader, opts Options) (*http.Response, error) {
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, errors.AddContext(err, fmt.Sprintf("could not create %v request", method))
	}
	if opts.APIKey != "" {
		req.SetBasicAuth("", opts.APIKey)
	}
	if opts.CustomUserAgent != "" {
		req.Header.Set("User-Agent", opts.CustomUserAgent)
	}
	if opts.customContentType != "" {
		req.Header.Set("Content-Type", opts.customContentType)
	}

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.AddContext(err, "could not execute request")
	}
	if resp.StatusCode >= 400 {
		return nil, errors.AddContext(makeResponseError(resp), "error code received")
	}

	return resp, nil
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

// makeURL makes a URL from the given parts.
func makeURL(portalURL, path string, query url.Values) string {
	url := fmt.Sprintf("%s/%s", strings.TrimRight(portalURL, "/"), strings.TrimLeft(path, "/"))
	if query != nil {
		url = fmt.Sprintf("%s?%s", url, query.Encode())
	}

	return url
}

// parseResponseBody parses the response body.
func parseResponseBody(resp *http.Response) (respBody *bytes.Buffer, err error) {
	defer func() {
		err = errors.Extend(err, resp.Body.Close())
	}()

	// parse the response
	respBody = &bytes.Buffer{}
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		return nil, errors.AddContext(err, "could not parse response body")
	}

	return respBody, nil
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
