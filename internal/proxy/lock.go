package proxy

import (
	"os"
	"path/filepath"
	"syscall"
)

const lockFileName = "emqutiti-proxy.lock"

// LockPath returns the location of the proxy lock file. All processes should
// use this path to coordinate proxy startup.
func LockPath() string {
	return filepath.Join(os.TempDir(), lockFileName)
}

// Acquire obtains an exclusive lock on the given file path and returns the
// locked file. The caller must call Release to free the lock.
func Acquire(path string) (*os.File, error) {
	return acquireLock(path)
}

// Release releases the lock by unlocking and removing the file. It is safe to
// call with a nil file.
func Release(f *os.File) error {
	if f == nil {
		return nil
	}
	name := f.Name()
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	if err := f.Close(); err != nil {
		return err
	}
	return os.Remove(name)
}

// acquireLock opens the file at the provided path and acquires an exclusive
// non-blocking lock using syscall.Flock. The caller should hold onto the
// returned file for the duration of the lock.
func acquireLock(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
