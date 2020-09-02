package skynet

import (
	"bytes"
	"encoding/json"
	"net/url"

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

	// AddSkykeyOptions contains the options used for addskykey.
	AddSkykeyOptions struct {
		Options
	}
	// CreateSkykeyOptions contains the options used for createskykey.
	CreateSkykeyOptions struct {
		Options
	}
	// GetSkykeyOptions contains the options used for skykey GET.
	GetSkykeyOptions struct {
		Options
	}
	// GetSkykeysOptions contains the options used for skykeys GET.
	GetSkykeysOptions struct {
		Options
	}

	// CreateSkykeyResponse contains the response for creating a skykey.
	CreateSkykeyResponse Skykey
	// GetSkykeyResponse contains the response for getting a skykey.
	GetSkykeyResponse Skykey
	// GetSkykeysResponse contains the response for listing skykeys.
	GetSkykeysResponse struct {
		// Skykeys is the returned list of skykeys.
		Skykeys []Skykey `json:"skykeys"`
	}
)

var (
	// DefaultAddSkykeyOptions contains the default addskykey options.
	DefaultAddSkykeyOptions = AddSkykeyOptions{
		Options: DefaultOptions("/skynet/addskykey"),
	}
	// DefaultCreateSkykeyOptions contains the default createskykey options.
	DefaultCreateSkykeyOptions = CreateSkykeyOptions{
		Options: DefaultOptions("/skynet/createskykey"),
	}
	// DefaultGetSkykeyOptions contains the default skykey GET options.
	DefaultGetSkykeyOptions = GetSkykeyOptions{
		Options: DefaultOptions("/skynet/skykey"),
	}
	// DefaultGetSkykeysOptions contains the default skykeys options.
	DefaultGetSkykeysOptions = GetSkykeysOptions{
		Options: DefaultOptions("/skynet/skykeys"),
	}
)

// AddSkykey stores the given base-64 encoded skykey with the skykey manager.
func (sc *SkynetClient) AddSkykey(skykey string, opts AddSkykeyOptions) error {
	body := &bytes.Buffer{}
	values := url.Values{}
	values.Set("skykey", skykey)

	_, err := sc.executeRequest(
		requestOptions{
			Options: opts.Options,
			method:  "POST",
			reqBody: body,
			query:   values,
		},
	)
	if err != nil {
		return errors.AddContext(err, "could not execute request")
	}

	return nil
}

// CreateSkykey returns a new skykey created and stored under the given name
// with the given type. skykeyType can be either "public-id" or "private-id".
func (sc *SkynetClient) CreateSkykey(name, skykeyType string, opts CreateSkykeyOptions) (Skykey, error) {
	body := &bytes.Buffer{}
	values := url.Values{}
	values.Set("name", name)
	values.Set("type", skykeyType)

	resp, err := sc.executeRequest(
		requestOptions{
			Options: opts.Options,
			method:  "POST",
			reqBody: body,
			query:   values,
		},
	)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not execute request")
	}

	respBody, err := parseResponseBody(resp)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not parse response body")
	}

	var apiResponse CreateSkykeyResponse
	err = json.Unmarshal(respBody.Bytes(), &apiResponse)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not unmarshal response JSON")
	}

	return Skykey(apiResponse), nil
}

// GetSkykeyByName returns the given skykey given its name.
func (sc *SkynetClient) GetSkykeyByName(name string, opts GetSkykeyOptions) (Skykey, error) {
	body := &bytes.Buffer{}
	values := url.Values{}
	values.Set("name", name)

	resp, err := sc.executeRequest(
		requestOptions{
			Options: opts.Options,
			method:  "GET",
			reqBody: body,
			query:   values,
		},
	)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not execute request")
	}

	respBody, err := parseResponseBody(resp)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not parse response body")
	}

	var apiResponse GetSkykeyResponse
	err = json.Unmarshal(respBody.Bytes(), &apiResponse)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not unmarshal response JSON")
	}

	return Skykey(apiResponse), nil
}

// GetSkykeyByID returns the given skykey given its ID.
func (sc *SkynetClient) GetSkykeyByID(id string, opts GetSkykeyOptions) (Skykey, error) {
	body := &bytes.Buffer{}
	values := url.Values{}
	values.Set("id", id)

	resp, err := sc.executeRequest(
		requestOptions{
			Options: opts.Options,
			method:  "GET",
			reqBody: body,
			query:   values,
		},
	)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not execute request")
	}

	respBody, err := parseResponseBody(resp)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not parse response body")
	}

	var apiResponse GetSkykeyResponse
	err = json.Unmarshal(respBody.Bytes(), &apiResponse)
	if err != nil {
		return Skykey{}, errors.AddContext(err, "could not unmarshal response JSON")
	}

	return Skykey(apiResponse), nil
}

// GetSkykeys returns a list of all skykeys.
func (sc *SkynetClient) GetSkykeys(opts GetSkykeysOptions) ([]Skykey, error) {
	resp, err := sc.executeRequest(
		requestOptions{
			Options: opts.Options,
			method:  "GET",
			reqBody: &bytes.Buffer{},
		},
	)
	if err != nil {
		return nil, errors.AddContext(err, "could not execute request")
	}

	respBody, err := parseResponseBody(resp)
	if err != nil {
		return nil, errors.AddContext(err, "could not parse response body")
	}

	var apiResponse GetSkykeysResponse
	err = json.Unmarshal(respBody.Bytes(), &apiResponse)
	if err != nil {
		return nil, errors.AddContext(err, "could not unmarshal response JSON")
	}

	return apiResponse.Skykeys, nil
}
