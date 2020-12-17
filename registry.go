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
	RegistryEntryResponse struct {
		Data      string `json:"data"`
		Revision  uint64 `json:"revision"`
		Signature string `json:"signature"`
	}

	RegistryEntry struct {
		DataKey  string
		Data     string
		Revision uint64
	}

	SignedEntry struct {
		Entry     RegistryEntry
		Signature []byte
	}

	SetEntryPublicKey struct {
		Algorithm string `json:"algorithm"`
		Key       []int  `json:"key"`
	}

	SetEntryBody struct {
		Publickey SetEntryPublicKey `json:"publickey"`
		Datakey   string            `json:"datakey"`
		Revision  int               `json:"revision"`
		Data      []int             `json:"data"`
		Signature []int             `json:"signature"`
	}
)

func (sc *SkynetClient) GetEntry(
	publicKey string,
	dataKey string,
) (r RegistryEntryResponse, err error) {
	// TODO: use timeout
	dataKeyHash := hashDataKey(dataKey)

	values := url.Values{}
	values.Set("publickey", fmt.Sprintf("ed25519:%s", publicKey))
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

	signedEntry := SignedEntry{
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

	requestBody := SetEntryBody{
		Publickey: SetEntryPublicKey{
			Algorithm: "ed25519",
			Key:       publicKeyBufferArray,
		},
		Datakey:   hex.EncodeToString(hashDataKey(entry.DataKey)),
		Revision:  int(entry.Revision),
		Data:      entryDataArray,
		Signature: signatureArray,
	}

	return json.Marshal(requestBody)
}
