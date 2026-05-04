package service

import (
	"sync"
)

type MemoryVaultSession struct {
	mu        sync.RWMutex
	masterKey []byte
}

func NewMemoryVaultSession() *MemoryVaultSession {
	return &MemoryVaultSession{}
}

func (s *MemoryVaultSession) SetMasterKey(key []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.masterKey = append([]byte(nil), key...)
}

func (s *MemoryVaultSession) GetMasterKey() ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.masterKey) == 0 {
		return nil, false
	}
	return append([]byte(nil), s.masterKey...), true
}

func (s *MemoryVaultSession) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.masterKey {
		s.masterKey[i] = 0
	}

	s.masterKey = nil
}

func (s *MemoryVaultSession) IsUnlocked() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.masterKey) > 0
}
