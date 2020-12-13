package skynet

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gitlab.com/NebulousLabs/errors"
	"golang.org/x/crypto/ed25519"
	"net/http"
	"net/url"
)

const registryEndpoint = "/skynet/registry"

type (
	RegistryEntry struct {
		Data      string `json:"data"`
		Revision  int64  `json:"revision"`
		Signature string `json:"signature"`
	}

	Entry struct {
		DataKey  string
		Data     string
		Revision int64
	}

	SignedEntry struct {
		Entry     Entry
		Signature []byte
	}

	GetEntryOptions struct {
		Timeout int64
	}
)

var (
	DefaultGetEntryOptions = GetEntryOptions{
		Timeout: 5000,
	}
)

func (sc *SkynetClient) GetEntry(
	publicKey string,
	dataKey string,
	_ GetEntryOptions,
) (r RegistryEntry, err error) {
	// TODO: use timeout
	hash, err := hashDataKey(dataKey)

	values := url.Values{}
	values.Set("publickey", fmt.Sprintf("ed25519:%s", publicKey))
	values.Set("datakey", hex.EncodeToString(hash))

	resp, err := sc.executeRequest(
		requestOptions{
			Options:   sc.Options,
			method:    "GET",
			reqBody:   &bytes.Buffer{},
			extraPath: registryEndpoint,
			query:     values,
		},
	)
	if err != nil {
		return r, errors.AddContext(err, "could not execute request")
	}

	defer func() {
		err2 := resp.Body.Close()
		if err != nil {
			err = errors.Compose(err, err2)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return r, errors.New("could not fetch registry entry")
	}

	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return r, errors.AddContext(err, "could not decode registry entry")
	}

	v, err := verifySignature(publicKey, dataKey, r)
	if err != nil {
		return r, err
	}

	if !v {
		return r, errors.New("could not verify signature from retrieved, signed registry entry -- possible corrupted entry")
	}

	return r, nil
}

func verifySignature(
	publicKey string,
	dataKey string,
	registryEntry RegistryEntry,
) (bool, error) {
	decodedSignature, err := hex.DecodeString(registryEntry.Signature)
	if err != nil {
		return false, errors.New("could not decode signature")
	}

	decodedData, err := hex.DecodeString(registryEntry.Data)
	if err != nil {
		return false, errors.New("could not decode data")
	}

	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return false, errors.New("could not decode publicKey")
	}

	signedEntry := SignedEntry{
		Entry: Entry{
			DataKey:  dataKey,
			Data:     string(decodedData),
			Revision: registryEntry.Revision,
		},
		Signature: decodedSignature,
	}

	message, err := hashRegistryEntry(signedEntry)
	if err != nil {
		return false, errors.AddContext(err, "could not hash registry entry")
	}

	return ed25519.Verify(publicKeyBytes, message, decodedSignature), nil
}
