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

const (
	registryEndpoint = "/skynet/registry"
	ed25519Algorithm = "ed25519"
)

type (
	// RegistryEntryResponse contains the response of the registry entry call.
	RegistryEntryResponse struct {
		// Data contains the stored data in the entry.
		Data string `json:"data"`
		// Revision is the revision number for the entry.
		Revision uint64 `json:"revision"`
		// Signature is the signature of the registry entry.
		Signature string `json:"signature"`
	}

	// RegistryEntry is the content of the registry entry.
	RegistryEntry struct {
		// DataKey is the key of the data for the given entry.
		DataKey string
		// Data contains the stored data in the entry.
		Data string
		// Revision is the revision number for the entry.
		Revision uint64
	}

	// SignedRegistryEntry is the signed registry entry.
	SignedRegistryEntry struct {
		// Entry is the content of the registry entry.
		Entry RegistryEntry
		// Signature is the signature of the registry entry.
		Signature []byte
	}

	// SetEntryPublicKey contains information about registry entry publicKey.
	SetEntryPublicKey struct {
		// Algorithm is the used algorithm
		Algorithm string `json:"algorithm"`
		// Key is the publicKey.
		Key []int `json:"key"`
	}

	// SetEntryRequestBody is the body content used to set registry entry.
	SetEntryRequestBody struct {
		// Publickey contains information about registry entry publicKey.
		Publickey SetEntryPublicKey `json:"publickey"`
		// DataKey is the key of the data for the given entry.
		Datakey string `json:"datakey"`
		// Revision is the revision number for the entry.
		Revision int `json:"revision"`
		// Data contains the stored data in the entry.
		Data []int `json:"data"`
		// Signature is the signature of the registry entry.
		Signature []int `json:"signature"`
	}
)

// GetEntry gets the registry entry corresponding to the publicKey and dataKey.
func (sc *SkynetClient) GetEntry(
	publicKey string,
	dataKey string,
) (r RegistryEntryResponse, err error) {
	// TODO: use timeout
	dataKeyHash := hashDataKey(dataKey)

	values := url.Values{}
	values.Set("publickey", fmt.Sprintf("%s:%s", ed25519Algorithm, publicKey))
	values.Set("datakey", hex.EncodeToString(dataKeyHash))

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

// verifySignature verifies signature from retrieved, signed registry entry
func verifySignature(
	publicKey string,
	dataKey string,
	registryEntry RegistryEntryResponse,
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

	signedEntry := SignedRegistryEntry{
		Entry: RegistryEntry{
			DataKey:  dataKey,
			Data:     string(decodedData),
			Revision: registryEntry.Revision,
		},
		Signature: decodedSignature,
	}

	return ed25519.Verify(
		publicKeyBytes,
		hashRegistryEntry(signedEntry.Entry),
		decodedSignature,
	), nil
}

// SetEntry sets the registry entry.
func (sc *SkynetClient) SetEntry(
	privateKey string,
	entry RegistryEntry,
) (err error) {
	requestBody, err := prepareSetEntryRequestBody(privateKey, entry)
	if err != nil {
		return errors.AddContext(err, "could not create request body")
	}

	options := sc.Options
	options.customContentType = "application/json"

	resp, err := sc.executeRequest(
		requestOptions{
			Options:   options,
			method:    "POST",
			reqBody:   bytes.NewBuffer(requestBody),
			extraPath: registryEndpoint,
		},
	)
	if err != nil {
		return errors.AddContext(err, "could not execute request")
	}

	defer func() {
		err2 := resp.Body.Close()
		if err != nil {
			err = errors.Compose(err, err2)
		}
	}()

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("could not set registry entry")
	}

	return nil
}

// prepareSetEntryRequestBody generates the body content used to set entry
func prepareSetEntryRequestBody(
	privateKey string,
	entry RegistryEntry,
) ([]byte, error) {
	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, errors.New("could not decode privateKey")
	}

	signature := ed25519.Sign(privateKeyBytes, hashRegistryEntry(entry))
	publicKeyBuffer := publicKeyFromPrivateKey(privateKeyBytes)

	publicKeyBufferArray, err := intSliceFromBytes(publicKeyBuffer)
	if err != nil {
		return nil, err
	}

	entryDataArray, err := intSliceFromBytes([]byte(entry.Data))
	if err != nil {
		return nil, err
	}

	signatureArray, err := intSliceFromBytes(signature)
	if err != nil {
		return nil, err
	}

	requestBody := SetEntryRequestBody{
		Publickey: SetEntryPublicKey{
			Algorithm: ed25519Algorithm,
			Key:       publicKeyBufferArray,
		},
		Datakey:   hex.EncodeToString(hashDataKey(entry.DataKey)),
		Revision:  int(entry.Revision),
		Data:      entryDataArray,
		Signature: signatureArray,
	}

	return json.Marshal(requestBody)
}
