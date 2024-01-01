package vbutton

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type FileSystemStorage struct {
	root string
}

func NewFileSystemStorage(root string) *FileSystemStorage {
	return &FileSystemStorage{root}
}

func (s *FileSystemStorage) Init() error {
	return os.MkdirAll(s.root, 0755)
}

func (s *FileSystemStorage) ServeFile(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.root, r.URL.Path))
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

func (s *FileSystemStorage) DeleteFile(name string) error {
	return os.Remove(filepath.Join(s.root, name))
}

func (s *FileSystemStorage) GetFile(name string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(s.root, name))
}
