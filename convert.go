package skynet

type (
	// ConvertOptions contains the options used for convert.
	ConvertOptions struct {
		Options
	}
)

var (
	// DefaultConvertOptions contains the default convert options.
	DefaultConvertOptions = ConvertOptions{
		Options: DefaultOptions("/skynet/skyfile"),
	}
)

// Convert converts an existing siafile to a skyfile and skylink.
func Convert(srcSiaPath, destSiaPath string, opts ConvertOptions) (string, error) {
	panic("Not implemented")
}
