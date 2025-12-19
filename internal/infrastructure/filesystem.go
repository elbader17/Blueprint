package infrastructure

import (
	"os"
)

// OSFileSystem implements domain.FileSystemPort using the os package
type OSFileSystem struct{}

func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

func (fs *OSFileSystem) MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

func (fs *OSFileSystem) WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func (fs *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (fs *OSFileSystem) CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

func (fs *OSFileSystem) Chmod(path string, mode uint32) error {
	return os.Chmod(path, os.FileMode(mode))
}

func (fs *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}
