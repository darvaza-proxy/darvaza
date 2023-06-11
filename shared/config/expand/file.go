// Package expand implements helpers to expand shell-style variables
// within config files before they are parsed
package expand

import (
	"io"
	"os"

	"mvdan.cc/sh/v3/shell"
)

// FromString consumes a string and expands shell-style variable
// within it. os.GetEnv will be used unless a custom mapper is provided
func FromString(s string, getEnv func(string) string) (string, error) {
	if getEnv == nil {
		getEnv = os.Getenv
	}

	if s == "" {
		// empty string
		return s, nil
	}

	return shell.Expand(s, getEnv)
}

// FromBytes consumes a byte array and expands shell-style variable
// within it. os.GetEnv will be used unless a custom mapper is provided
func FromBytes(b []byte, getEnv func(string) string) (string, error) {
	return FromString(string(b), getEnv)
}

// FromReader consumes a io.Reader and expands shell-style variable
// within it. os.GetEnv will be used unless a custom mapper is provided
func FromReader(f io.Reader, getEnv func(string) string) (string, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return FromString(string(b[:]), getEnv)
}

// FromFile reads a file and expands shell-style variable
// within it. os.GetEnv will be used unless a custom mapper is provided
func FromFile(filename string, getEnv func(string) string) (string, error) {
	var f io.Reader

	if filename == "-" {
		f = os.Stdin
	} else if file, err := os.Open(filename); err != nil {
		return "", err
	} else {
		defer file.Close()
		f = file
	}

	return FromReader(f, getEnv)
}
