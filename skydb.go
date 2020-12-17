package skynet

import (
	"encoding/hex"
	"fmt"
	"gitlab.com/NebulousLabs/errors"
	"io"
	"io/ioutil"
	"log"
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
) error {
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

	tmpFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("temp-%s", dataKey))
	if err != nil {
		log.Fatal("could not create temporary file", err)
	}

	if _, err = tmpFile.Write(json); err != nil {
		log.Fatal("failed to write to temporary file", err)
	}

	defer os.Remove(tmpFile.Name())

	skylink, err := sc.UploadFile(tmpFile.Name(), DefaultUploadOptions)
	if err != nil {
		return errors.New("could not upload file")
	}

	skylink = strings.TrimPrefix(skylink, URISkynetPrefix)

	return sc.SetEntry(privateKey, RegistryEntry{
		DataKey:  dataKey,
		Data:     skylink,
		Revision: *revision,
	})
}
