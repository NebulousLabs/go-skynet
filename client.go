package skynet

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"gitlab.com/NebulousLabs/errors"
)

type (
	// SkynetClient is the Skynet Client which can be used to access Skynet.
	SkynetClient struct {
		PortalURL string
		Options   Options
	}

	// requestOptions contains the options for a request.
	requestOptions struct {
		Options

		method    string
		reqBody   io.Reader
		extraPath string
		query     url.Values
	}
)

// New creates a new Skynet Client which can be used to access Skynet.
func New() SkynetClient {
	return NewCustom("", Options{})
}

// NewCustom creates a new Skynet Client with a custom portal URL and options.
// Pass in "" for the portal to let the function select one for you.
func NewCustom(portalURL string, customOptions Options) SkynetClient {
	if portalURL == "" {
		portalURL = DefaultPortalURL()
	}
	return SkynetClient{
		PortalURL: portalURL,
		Options:   customOptions,
	}
}

// executeRequest makes and executes a request.
func (sc *SkynetClient) executeRequest(config requestOptions) (*http.Response, error) {
	url := sc.PortalURL
	method := config.method
	reqBody := config.reqBody

	// Set options, prioritizing options passed to the API calls.
	opts := sc.Options
	if config.EndpointPath != "" {
		opts.EndpointPath = config.EndpointPath
	}
	if config.APIKey != "" {
		opts.APIKey = config.APIKey
	}
	if config.CustomUserAgent != "" {
		opts.CustomUserAgent = config.CustomUserAgent
	}
	if config.customContentType != "" {
		opts.customContentType = config.customContentType
	}

	// Make the URL.
	url = makeURL(url, opts.EndpointPath, config.extraPath, config.query)

	// Create the request.
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.AddContext(err, "could not execute request")
	}
	if resp.StatusCode >= 400 {
		return nil, errors.AddContext(makeResponseError(resp), "error code received")
	}

	return resp, nil
}
