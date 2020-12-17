package skynet

import (
	"encoding/hex"
	"fmt"
	"gitlab.com/NebulousLabs/errors"
	"io"
	"io/ioutil"
	"math"
	"os"
	"strings"
)

// GetJSON gets the JSON object corresponding to the publicKey and dataKey.
func (sc *SkynetClient) GetJSON(
	publicKey string,
	dataKey string,
) (io.ReadCloser, error) {
	entry, err := sc.GetEntry(publicKey, dataKey)
	if err != nil {
		return nil, errors.AddContext(err, "could not get entry")
	}

	fmt.Println(entry)

	skylink, err := hex.DecodeString(entry.Data)
	if err != nil {
		return nil, errors.New("could not decode data")
	}

	return sc.Download(string(skylink), DefaultDownloadOptions)
}

// SetJSON sets a JSON object at the registry entry corresponding to the publicKey and dataKey.
func (sc *SkynetClient) SetJSON(
	privateKey string,
	dataKey string,
	json io.Reader,
	revision *uint64,
) (err error) {
	if revision == nil {
		privateKeyBytes, err := hex.DecodeString(privateKey)
		if err != nil {
			return errors.AddContext(err, "could not decode privateKey")
		}

		publicKeyBytes, err := publicKeyFromPrivateKey(privateKeyBytes)
		if err != nil {
			return errors.AddContext(err, "could not get publicKey from privateKey")
		}

		entry, err := sc.GetEntry(hex.EncodeToString(publicKeyBytes), dataKey)
		if err != nil {
			return errors.AddContext(err, "could not get entry")
		}

		newRevision := entry.Revision + 1
		revision = &newRevision

		if newRevision > math.MaxUint64 {
			return errors.New("current entry already has maximum allowed revision, could not update the entry")
		}
	}

	tempFile, err := createTempFileFromJson(dataKey, json)
	if err != nil {
		return
	}

	if tempFile == nil {
		return errors.New("could not create temp JSON file")
	}

	defer func() {
		err = os.Remove(tempFile.Name())
		return
	}()

	skylink, err := sc.UploadFile(tempFile.Name(), DefaultUploadOptions)
	if err != nil {
		return errors.AddContext(err, "could not upload file")
	}

	skylink = strings.TrimPrefix(skylink, URISkynetPrefix)

	return sc.SetEntry(privateKey, RegistryEntry{
		DataKey:  dataKey,
		Data:     skylink,
		Revision: *revision,
	})
}

// createTempFileFromJson creates a temp file from a passed io.Reader
func createTempFileFromJson(filename string, json io.Reader) (f *os.File, err error) {
	if f, err = ioutil.TempFile(os.TempDir(), fmt.Sprintf("temp-%s", filename)); err != nil {
		return nil, errors.AddContext(err, "could not create temp file")
	}

	reader := io.TeeReader(json, f)
	if _, err = ioutil.ReadAll(reader); err != nil {
		return nil, errors.AddContext(err, "could not write JSON to temp file")
	}

	return
}
