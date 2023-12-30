package vbutton

import (
	"io"
	"os"
	"path/filepath"
)

type FileSystemStorage struct {
	root string
}

func NewFileSystemStorage(root string) *FileSystemStorage {
	return &FileSystemStorage{root}
}

func (s *FileSystemStorage) SaveFile(name string, content io.Reader) error {
	f, err := os.Create(filepath.Join(s.root, name))

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, content)

	return err
}

func (s *FileSystemStorage) GetFile(name string) (io.Reader, error) {
	return os.Open(filepath.Join(s.root, name))
}
