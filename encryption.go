package skynet

import (
	"bytes"
	"encoding/json"
	"fmt"

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

	// EncryptionOptions contains options used for encrypting uploads and
	// decrypting downloads.
	EncryptionOptions struct {
		// SkykeyName is the name of the skykey used to encrypt the upload.
		SkykeyName string
		// SkykeyID is the ID of the skykey used to encrypt the upload.
		SkykeyID string
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
	// ListSkykeysOptions contains the options used for skykeys GET.
	ListSkykeysOptions struct {
		Options
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
	// DefaultListSkykeysOptions contains the default skykeys options.
	DefaultListSkykeysOptions = ListSkykeysOptions{
		Options: DefaultOptions("/skynet/skykeys"),
	}
)

// AddSkykey stores the given base-64 encoded skykey with the skykey manager.
func AddSkykey(skykey string, opts AddSkykeyOptions) error {
	body := &bytes.Buffer{}
	url := makeURL(opts.PortalURL, opts.EndpointPath)
	url = fmt.Sprintf("%s?skykey=%s", url, skykey)

	_, err := executeRequest(opts.Options, "POST", url, body)
	if err != nil {
		return errors.AddContext(err, "could not execute request")
	}

	return nil
}

// CreateSkykey returns a new skykey created and stored under the given name
// with the given type. skykeyType can be either "public-id" or "private-id".
func CreateSkykey(name, skykeyType string, opts CreateSkykeyOptions) (Skykey, error) {
	body := &bytes.Buffer{}
	url := makeURL(opts.PortalURL, opts.EndpointPath)
	url = fmt.Sprintf("%s?name=%s&type=%s", url, name, skykeyType)

	resp, err := executeRequest(opts.Options, "POST", url, body)
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
func GetSkykeyByName(name string, opts GetSkykeyOptions) (Skykey, error) {
	body := &bytes.Buffer{}
	url := makeURL(opts.PortalURL, opts.EndpointPath)
	url = fmt.Sprintf("%s?name=%s", url, name)

	resp, err := executeRequest(opts.Options, "GET", url, body)
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
func GetSkykeyByID(id string, opts GetSkykeyOptions) (Skykey, error) {
	body := &bytes.Buffer{}
	url := makeURL(opts.PortalURL, opts.EndpointPath)
	url = fmt.Sprintf("%s?id=%s", url, id)

	resp, err := executeRequest(opts.Options, "GET", url, body)
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

// ListSkykeys returns a list of all skykeys.
func ListSkykeys(opts ListSkykeysOptions) ([]Skykey, error) {
	url := makeURL(opts.PortalURL, opts.EndpointPath)

	resp, err := executeRequest(opts.Options, "GET", url, &bytes.Buffer{})
	if err != nil {
		return nil, errors.AddContext(err, "could not execute request")
	}

	respBody, err := parseResponseBody(resp)
	if err != nil {
		return nil, errors.AddContext(err, "could not parse response body")
	}

	var apiResponse ListSkykeysResponse
	err = json.Unmarshal(respBody.Bytes(), &apiResponse)
	if err != nil {
		return nil, errors.AddContext(err, "could not unmarshal response JSON")
	}

	return apiResponse.Skykeys, nil
}
