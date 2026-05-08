package confy

import (
	"io"
	"os"
)

type FS interface {
	Open(name string) (io.ReadCloser, error)
	Stat(name string) (os.FileInfo, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
}

type osFS struct{}

func (osFS) Open(name string) (io.ReadCloser, error) { return os.Open(name) }
func (osFS) Stat(name string) (os.FileInfo, error)   { return os.Stat(name) }
func (osFS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}
