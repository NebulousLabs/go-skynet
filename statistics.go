package skynet

type (
	// GetStatisticsOptions contains the options used for get statistics.
	GetStatisticsOptions struct {
		Options
	}
)

var (
	// DefaultGetStatisticsOptions contains the default get statistics options.
	DefaultGetStatisticsOptions = GetStatisticsOptions{
		Options: DefaultOptions("/skynet/stats"),
	}
)

// GetStatistics returns statistical information about Skynet, e.g. number of
// files uplaoded.
func GetStatistics(opts GetStatisticsOptions) error {
	panic("Not implemented")
}
