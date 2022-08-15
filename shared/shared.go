package darvaza

// Reloader identifies a service or worker that allows
// its internal configuration to be reloaded on runtime
type Reloader interface {
	Reload() error
}
