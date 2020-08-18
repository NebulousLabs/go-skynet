package skynet

type (
	// GetBlocklistOptions contains the options used for get blocklist.
	GetBlocklistOptions struct {
		Options
	}

	// UpdateBlocklistOptions contains the options used for update blocklist.
	UpdateBlocklistOptions struct {
		Options
	}
)

var (
	// DefaultGetBlocklistOptions contains the default get blocklist options.
	DefaultGetBlocklistOptions = GetBlocklistOptions{
		Options: DefaultOptions("/skynet/blocklist"),
	}

	// DefaultUpdateBlocklistOptions contains the default update blocklist
	// options.
	DefaultUpdateBlocklistOptions = GetBlocklistOptions{
		Options: DefaultOptions("/skynet/blocklist"),
	}
)

// GetBlocklist returns the list of hashed merkleroots that are blocklisted.
func (sc *SkynetClient) GetBlocklist(opts GetBlocklistOptions) ([]string, error) {
	panic("Not implemented")
}

// UpdateBlocklist updates the list of skylinks that should be blocklisted from
// Skynet. This function can be used to both add and remove skylinks from the
// blocklist.
func (sc *SkynetClient) UpdateBlocklist(additions, removals []string, opts UpdateBlocklistOptions) error {
	panic("Not implemented")
}
