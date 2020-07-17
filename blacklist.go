package skynet

type (
	// GetBlacklistOptions contains the options used for get blacklist.
	GetBlacklistOptions struct {
		Options
	}

	// UpdateBlacklistOptions contains the options used for update blacklist.
	UpdateBlacklistOptions struct {
		Options
	}
)

var (
	// DefaultGetBlacklistOptions contains the default get blacklist options.
	DefaultGetBlacklistOptions = GetBlacklistOptions{
		Options: DefaultOptions("/skynet/blacklist"),
	}

	// DefaultUpdateBlacklistOptions contains the default update blacklist
	// options.
	DefaultUpdateBlacklistOptions = GetBlacklistOptions{
		Options: DefaultOptions("/skynet/blacklist"),
	}
)

// GetBlacklist returns the list of hashed merkleroots that are blacklisted.
func GetBlacklist(opts GetBlacklistOptions) ([]string, error) {
	panic("Not implemented")
}

// UpdateBlacklist updates the list of skylinks that should be blacklisted from
// Skynet. This function can be used to both add and remove skylinks from the
// blacklist.
func UpdateBlacklist(additions, removals []string, opts UpdateBlacklistOptions) error {
	panic("Not implemented")
}
