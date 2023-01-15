// Package version exists so the linker can inject Version and BuildData values
package version

var (
	// Version contains the git hashtag injected by make
	Version = "N/A"
	// BuildDate contains the build timestamp injected by make
	BuildDate = "N/A"
)
