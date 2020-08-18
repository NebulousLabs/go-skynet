package skynet

type (
	// PinOptions contains the options used for pin.
	PinOptions struct {
		Options
	}

	// UnpinOptions contains the options used for unpin.
	UnpinOptions struct {
		Options
		EndpointPathUnpinDir  string
		EndpointPathUnpinFile string
	}
)

var (
	// DefaultPinOptions contains the default pin options
	DefaultPinOptions = PinOptions{
		Options: DefaultOptions("/skynet/pin"),
	}

	// DefaultUnpinOptions contains the default unpin options.
	DefaultUnpinOptions = UnpinOptions{
		Options:               DefaultOptions(""),
		EndpointPathUnpinDir:  "/renter/dir",
		EndpointPathUnpinFile: "/renter/delete",
	}
)

// Pin pins the file associated with this skylink by re-uploading an exact copy.
func (sc *SkynetClient) Pin(skylink, destSiaPath string, opts PinOptions) error {
	panic("Not implemented")
}

// Unpin unpins the pinned skyfile or directory at the given siapath.
func (sc *SkynetClient) Unpin(siaPath string, opts UnpinOptions) error {
	panic("Not implemented")
}
