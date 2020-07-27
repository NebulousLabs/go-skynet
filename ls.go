package skynet

type (
	// LsOptions contains the options used for ls.
	LsOptions struct {
		Options
		EndpointPathLsDir  string
		EndpointPathLsFile string
	}
)

var (
	// DefaultLsOptions conains the default ls options.
	DefaultLsOptions = LsOptions{
		Options:            DefaultOptions(""),
		EndpointPathLsDir:  "/renter/dir",
		EndpointPathLsFile: "/renter/file",
	}
)

// Ls returns the list of files and/or directories at the given path.
func Ls(siaPath string, opts LsOptions) error {
	panic("Not implemented")
}
