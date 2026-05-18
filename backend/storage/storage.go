package storage

import (
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Store interface {
	Save(challengeID, userID string, r io.Reader) (string, error)
	BaseDir() string
}

type LocalStore struct {
	base string
}

func NewLocalStore() *LocalStore {
	base := os.Getenv("DATA_DIR")
	if base == "" {
		base = "./data/videos"
	}
	return &LocalStore{base: base}
}

func (s *LocalStore) Save(challengeID, userID string, r io.Reader) (string, error) {
	dir := filepath.Join(s.base, challengeID, userID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	id := uuid.New().String()
	dst := filepath.Join(dir, id+".mp4")

	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}

	return dst, nil
}

func (s *LocalStore) BaseDir() string {
	return s.base
}
