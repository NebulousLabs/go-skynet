package skynet

type (
	// ListFilesOptions contains the options used for list files.
	ListFilesOptions struct {
		Options
		EndpointPathListFilesDir  string
		EndpointPathListFilesFile string
	}
)

var (
	// DefaultListFilesOptions conains the default list files options.
	DefaultListFilesOptions = ListFilesOptions{
		Options:                   DefaultOptions(""),
		EndpointPathListFilesDir:  "/renter/dir",
		EndpointPathListFilesFile: "/renter/file",
	}
)

// ListFiles returns the list of files and/or directories at the given path.
func (sc *SkynetClient) ListFiles(siaPath string, opts ListFilesOptions) error {
	panic("Not implemented")
}
