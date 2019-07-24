package client

// AsuraQueryOptions can be used to provide options for AsuraQuery call other
// than the DefaultAsuraQueryOptions.
type AsuraQueryOptions struct {
	Height int64
	Prove  bool
}

// DefaultAsuraQueryOptions are latest height (0) and prove false.
var DefaultAsuraQueryOptions = AsuraQueryOptions{Height: 0, Prove: false}
