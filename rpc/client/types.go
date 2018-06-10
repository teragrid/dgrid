package client

// AsuraQueryOptions can be used to provide options for AsuraQuery call other
// than the DefaultAsuraQueryOptions.
type AsuraQueryOptions struct {
	Height  int64
	Trusted bool
}

// DefaultAsuraQueryOptions are latest height (0) and trusted equal to false
// (which will result in a proof being returned).
var DefaultAsuraQueryOptions = AsuraQueryOptions{Height: 0, Trusted: false}
