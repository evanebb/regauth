package client

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type CredentialStore interface {
	// Save saves the credentials for the specified host.
	Save(host string, credentials HostCredentials) error
	// Delete will delete the currently stored credentials for the given host.
	Delete(host string) error
	// GetAll will get all hosts and their corresponding credentials.
	GetAll() (map[string]HostCredentials, error)
	// GetCurrent will get the currently selected host and the corresponding credentials.
	GetCurrent() (string, HostCredentials, error)
	// UseHost will set the current host and credentials to use for requests.
	UseHost(host string) error
}

type AuthConfig struct {
	Current     string                     `json:"current"`
	Credentials map[string]HostCredentials `json:"credentials"`
}

type HostCredentials struct {
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

// readAuthConfigFile will read and decode the local authentication configuration file.
//
// TODO: we currently read/write the file on every method call, it's probably better to read/write it once and hold it in memory
func (s FileCredentialStore) readAuthConfigFile() (AuthConfig, error) {
	emptyAuthConf := AuthConfig{Credentials: make(map[string]HostCredentials)}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// if we don't have an existing config file, return an empty configuration
			return emptyAuthConf, nil
		}

		return emptyAuthConf, err
	}

	var conf AuthConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		// if we encounter an error here, just return an empty configuration
		return emptyAuthConf, nil
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

func (s FileCredentialStore) Save(host string, credentials HostCredentials) error {
	auth, err := s.readAuthConfigFile()
	if err != nil {
		return err
	}

	auth.Credentials[host] = credentials
	auth.Current = host
	return s.writeAuthConfigFile(auth)
}

func (s FileCredentialStore) Delete(host string) error {
	auth, err := s.readAuthConfigFile()
	if err != nil {
		return err
	}

	delete(auth.Credentials, host)
	if auth.Current == host {
		auth.Current = ""
	}
	return s.writeAuthConfigFile(auth)
}

func (s FileCredentialStore) GetAll() (map[string]HostCredentials, error) {
	auth, err := s.readAuthConfigFile()
	if err != nil {
		return nil, err
	}

	return auth.Credentials, nil
}

func (s FileCredentialStore) GetCurrent() (string, HostCredentials, error) {
	auth, err := s.readAuthConfigFile()
	if err != nil {
		return "", HostCredentials{}, err
	}

	return auth.Current, auth.Credentials[auth.Current], err
}

func (s FileCredentialStore) UseHost(host string) error {
	auth, err := s.readAuthConfigFile()
	if err != nil {
		return err
	}

	if _, ok := auth.Credentials[host]; !ok {
		return errors.New("no credentials registered for given host")
	}

	auth.Current = host
	return s.writeAuthConfigFile(auth)
}
