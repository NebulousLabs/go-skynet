package skynet

type (
	// GetPortalsOptions contains the options used for get portals.
	GetPortalsOptions struct {
		Options
	}

	// UpdatePortalsOptions contains the options used for update portals.
	UpdatePortalsOptions struct {
		Options
	}

	// Portal contains information about a known portal.
	Portal struct {
		// Address is the IP or domain name and the port of the portal. Must be
		// a valid network address.
		Address string
		// Public indicates whether the portal can be accessed publicly or not.
		Public bool
	}
)

var (
	// DefaultGetPortalsOptions contains the default get portals options.
	DefaultGetPortalsOptions = GetPortalsOptions{
		Options: DefaultOptions("/skynet/portals"),
	}

	// DefaultUpdatePortalsOptions contains the default update portals
	// options.
	DefaultUpdatePortalsOptions = GetPortalsOptions{
		Options: DefaultOptions("/skynet/portals"),
	}
)

// GetPortals returns the list of known Skynet portals.
func (sc *SkynetClient) GetPortals(opts GetPortalsOptions) ([]Portal, error) {
	panic("Not implemented")
}

// UpdatePortals updates the list of known portals. This function can be used to
// both add and remove portals from the list. Removals are provided in the form
// of addresses.
func (sc *SkynetClient) UpdatePortals(additions []Portal, removals []string, opts UpdatePortalsOptions) error {
	panic("Not implemented")
}
