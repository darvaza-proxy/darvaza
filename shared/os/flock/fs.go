package flock

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"darvaza.org/x/fs/fssyscall"
)

const (
	// DefaultFileMode is the filemode used when a flock file is first created
	DefaultFileMode = 0666
	// DefaultDirMode is the filemode used when a directory is created
	DefaultDirMode = 0777
)

var (
	separator   = filepath.Separator
	slash       = string([]rune{separator})
	dotSlash    = string([]rune{'.', separator})
	dotDotSlash = string([]rune{'.', '.', separator})
)

// coalesceMode picks the first non-zero fs.FileMode of a list
func coalesceMode(modes ...fs.FileMode) fs.FileMode {
	for _, mode := range modes {
		if mode != 0 {
			return mode
		}
	}
	return 0
}

// Options provides rules for Flock protected actions
type Options struct {
	Base    string      // optional prefix
	Create  bool        // create nodes if missing
	DirMode fs.FileMode // mode used for MkdirAll
}

// JoinName joins the given name with the Options.Base when not qualified
func (opt Options) JoinName(name string) string {
	switch {
	case name == "":
		// empty
		return opt.Base
	case opt.Base == "",
		strings.HasPrefix(name, slash),
		strings.HasPrefix(name, dotSlash),
		strings.HasPrefix(name, dotDotSlash):
		// ignore Base
		return filepath.Clean(name)
	default:
		return filepath.Join(opt.Base, filepath.Clean(name))
	}
}

//revive:disable:confusing-results

// NameSplit considers Options.Base when splitting a given path
func (opt Options) NameSplit(name string) (string, string) {
	//revive:enable:confusing-results
	if name == "" {
		// empty
		return opt.Base, ""
	}

	// join with Base first
	return filepath.Split(opt.JoinName(name))
}

// NewOpener creates an opener funcion considering Options.Create and the given permissions
func (opt Options) NewOpener(perm fs.FileMode) func(string) (Handle, error) {
	perm = coalesceMode(perm, DefaultFileMode)

	fn := func(path string) (Handle, error) {
		mode := os.O_RDONLY
		if opt.Create {
			mode |= os.O_CREATE
		}
		return fssyscall.Open(path, mode, perm)
	}

	return fn
}

func (opt Options) newFileLock(name string, perm fs.FileMode) (*Flock, error) {
	dir, file := opt.NameSplit(name)
	if file == "" {
		return nil, syscall.EINVAL
	} else if dir != "" && opt.Create {
		err := mkdirAllCoalesceMode(dir, opt.DirMode, DefaultDirMode)
		if err != nil {
			return nil, err
		}
	}

	fl := NewWithOpener(filepath.Join(dir, file), opt.NewOpener(perm))
	return fl, nil
}

func (opt Options) newDirLock(name string, dmode fs.FileMode) (*Flock, error) {
	name = opt.JoinName(name)

	if _, err := os.Stat(name); err != nil {
		// failed
		if os.IsNotExist(err) && opt.Create {
			err = mkdirAllCoalesceMode(opt.Base, dmode, opt.DirMode, DefaultDirMode)
		}

		if err != nil {
			return nil, err
		}
	}

	fl := NewWithOpener(name, nil)
	return fl, nil
}

// New creates a new Flock considering Options
func (opt Options) New(name string, perm fs.FileMode) (*Flock, error) {
	return opt.newFileLock(name, perm)
}

// MkdirBase creates the Options.Base if it doesn't exist
func (opt Options) MkdirBase(dmode fs.FileMode) error {
	if opt.Base != "" {
		return mkdirAllCoalesceMode(opt.Base, dmode, opt.DirMode, DefaultDirMode)
	}
	return nil
}

// Mkdir creates a directory within the base
func (opt Options) Mkdir(name string, dmode fs.FileMode) error {
	return mkdirCoalesceMode(opt.JoinName(name), dmode, opt.DirMode, DefaultDirMode)
}

func mkdirCoalesceMode(fullname string, dmode ...fs.FileMode) error {
	return os.Mkdir(fullname, coalesceMode(dmode...))
}

// MkdirAll attempts to create all directories on a path within the base
func (opt Options) MkdirAll(name string, dmode fs.FileMode) error {
	return mkdirAllCoalesceMode(opt.JoinName(name), dmode, opt.DirMode, DefaultDirMode)
}

func mkdirAllCoalesceMode(fullname string, dmode ...fs.FileMode) error {
	return os.MkdirAll(fullname, coalesceMode(dmode...))
}

// ReadDir reads the entries of a directory, flocked
func (opt Options) ReadDir(name string) ([]fs.DirEntry, error) {
	fl, err := opt.newDirLock(name, 0)
	if err != nil {
		return nil, err
	}

	err = fl.Lock()
	if err != nil {
		return nil, err
	}
	defer fl.Unlock()

	return os.ReadDir(name)
}

// ReadFile reads a file whole, flocked
func (opt Options) ReadFile(name string, perm fs.FileMode) ([]byte, error) {
	perm = coalesceMode(perm, DefaultFileMode)

	fl, err := opt.newFileLock(name, perm)
	if err != nil {
		return nil, err
	}

	err = fl.Lock()
	if err != nil {
		return nil, err
	}
	defer fl.Unlock()

	return os.ReadFile(name)
}

// WriteFile writes the given content to a locked file
func (opt Options) WriteFile(name string, data []byte, perm fs.FileMode) error {
	perm = coalesceMode(perm, DefaultFileMode)

	fl, err := opt.newFileLock(name, perm)
	if err != nil {
		return err
	}

	err = fl.Lock()
	if err != nil {
		return err
	}
	defer fl.Unlock()

	return os.WriteFile(name, data, perm)
}
