package skynet

type (
	// GetStatsOptions contains the options used for get stats.
	GetStatsOptions struct {
		Options
	}
)

var (
	// DefaultGetStatsOptions contains the default get stats options.
	DefaultGetStatsOptions = GetStatsOptions{
		Options: DefaultOptions("/skynet/stats"),
	}
)

// GetStats returns statistical information about Skynet, e.g. number of
// files uplaoded.
func GetStats(opts GetStatsOptions) error {
	panic("Not implemented")
}
