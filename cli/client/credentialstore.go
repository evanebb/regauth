package client

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type CredentialStore interface {
	Save(credentials Credentials) error
	Get() (Credentials, error)
}

type AuthConfig struct {
	Credentials Credentials `json:"credentials"`
}

type Credentials struct {
	Host     string `json:"host,omitempty"`
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type FileCredentialStore struct {
	path string
}

func NewFileCredentialStore(path string) (FileCredentialStore, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return FileCredentialStore{}, err
	}

	return FileCredentialStore{path: path}, nil
}

func (s FileCredentialStore) readAuthConfigFile() (AuthConfig, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// if we don't have an existing config file, return an empty configuration
			return AuthConfig{}, nil
		}

		return AuthConfig{}, err
	}

	var conf AuthConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		// if we encounter an error here, just return an empty configuration
		return AuthConfig{}, nil
	}

	return conf, nil
}

func (s FileCredentialStore) writeAuthConfigFile(a AuthConfig) error {
	encoded, err := json.Marshal(a)
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, encoded, 0o600)
}

func (s FileCredentialStore) Save(credentials Credentials) error {
	auth, err := s.readAuthConfigFile()
	if err != nil {
		return err
	}

	auth.Credentials = credentials
	return s.writeAuthConfigFile(auth)
}

func (s FileCredentialStore) Get() (Credentials, error) {
	auth, err := s.readAuthConfigFile()
	if err != nil {
		return Credentials{}, err
	}

	return auth.Credentials, nil
}
