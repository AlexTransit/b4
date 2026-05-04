package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/daniellavrushin/b4/log"
)

const SecretsFileName = "ai_secrets.json"

type SecretStore struct {
	path string
	mu   sync.RWMutex
	data map[string]string
}

func NewSecretStore(configPath string) *SecretStore {
	dir := filepath.Dir(configPath)
	if dir == "" || dir == "." {
		dir = "."
	}
	s := &SecretStore{
		path: filepath.Join(dir, SecretsFileName),
		data: map[string]string{},
	}
	s.load()
	return s
}

func (s *SecretStore) Path() string {
	return s.path
}

func (s *SecretStore) load() {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Errorf("ai: failed to read %s: %v", s.path, err)
		}
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := json.Unmarshal(raw, &s.data); err != nil {
		log.Errorf("ai: failed to parse %s: %v", s.path, err)
		s.data = map[string]string{}
	}
}

func (s *SecretStore) save() error {
	s.mu.RLock()
	raw, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	if dir := filepath.Dir(s.path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *SecretStore) Get(ref string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[ref]
}

func (s *SecretStore) Set(ref, key string) error {
	if ref == "" {
		return nil
	}
	s.mu.Lock()
	if key == "" {
		delete(s.data, ref)
	} else {
		s.data[ref] = key
	}
	s.mu.Unlock()
	return s.save()
}

func (s *SecretStore) Delete(ref string) error {
	return s.Set(ref, "")
}

func (s *SecretStore) Refs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	refs := make([]string, 0, len(s.data))
	for k := range s.data {
		refs = append(refs, k)
	}
	return refs
}

func (s *SecretStore) Has(ref string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[ref]
	return ok
}
