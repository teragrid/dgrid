package client

// asuraQueryOptions can be used to provide options for asuraQuery call other
// than the DefaultasuraQueryOptions.
type asuraQueryOptions struct {
	Height  int64
	Trusted bool
}

// DefaultasuraQueryOptions are latest height (0) and trusted equal to false
// (which will result in a proof being returned).
var DefaultasuraQueryOptions = asuraQueryOptions{Height: 0, Trusted: false}
