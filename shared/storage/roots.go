package storage

// RootStore represents a view of the Store that can be used for generic setup
type RootStore interface {
	Store

	// AddCACert can either add a certificate chain by content, filename, or directory
	// to scan
	AddCACert(string) error

	// Prepare finishes the configuration of the store and makes sure it's in usable
	// state
	Prepare() error
}
