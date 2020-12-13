package skynet

import (
	"encoding/hex"
	"gitlab.com/NebulousLabs/errors"
	"io"
)

func (sc *SkynetClient) GetJson(
	publicKey string,
	dataKey string,
) (io.ReadCloser, error) {
	entry, err := sc.GetEntry(publicKey, dataKey, DefaultGetEntryOptions)
	if err != nil {
		return nil, errors.AddContext(err, "could not get entry")
	}

	skylink, err := hex.DecodeString(entry.Data)
	if err != nil {
		return nil, errors.New("could not decode data")
	}

	return sc.Download(string(skylink), DefaultDownloadOptions)
}
