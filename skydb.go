package skynet

import (
	"encoding/hex"
	"fmt"
	"gitlab.com/NebulousLabs/errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

const maxRevision uint64 = 18446744073709551615

func (sc *SkynetClient) GetJson(
	publicKey string,
	dataKey string,
) (io.ReadCloser, error) {
	entry, err := sc.GetEntry(publicKey, dataKey)
	if err != nil {
		return nil, errors.AddContext(err, "could not get entry")
	}

	skylink, err := hex.DecodeString(entry.Data)
	if err != nil {
		return nil, errors.New("could not decode data")
	}

	return sc.Download(string(skylink), DefaultDownloadOptions)
}

func (sc *SkynetClient) SetJson(
	privateKey string,
	dataKey string,
	json []byte,
	revision *uint64,
) (err error) {
	if revision == nil {
		privateKeyBytes, err := hex.DecodeString(privateKey)
		if err != nil {
			return errors.AddContext(err, "could not decode privateKey")
		}

		publicKeyBytes := publicKeyFromPrivateKey(privateKeyBytes)
		entry, err := sc.GetEntry(hex.EncodeToString(publicKeyBytes), dataKey)
		if err != nil {
			return errors.AddContext(err, "could not get entry")
		}

		newRevision := entry.Revision + 1
		revision = &newRevision

		if newRevision > maxRevision {
			return errors.New("current entry already has maximum allowed revision, could not update the entry")
		}
	}

	tempFile, err := createTempFileFromJson(dataKey, json
	if err != nil {
		return
	}

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

func createTempFileFromJson(filename string, json []byte) (f *os.File, err error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("temp-%s", filename))
	if err != nil {
		return nil, errors.AddContext(err, "could not create temp file")
	}

	if _, err = tmpFile.Write(json); err != nil {
		return nil, errors.AddContext(err, "could not write json to temp file")
	}

	defer func() {
		err = os.Remove(tmpFile.Name())
		return
	}()

	return
}
