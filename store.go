package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

type pairEntry struct {
	telegramID     int64
	expiresAt      time.Time
	failedAttempts int
}

type lookupStatus int

const (
	lookupNotFound lookupStatus = iota
	lookupOK
	lookupExpired
)

type pairStore struct {
	mu      sync.RWMutex
	entries map[string]*pairEntry
}

func newPairStore() *pairStore {
	return &pairStore{entries: make(map[string]*pairEntry)}
}

func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (s *pairStore) create(telegramID int64, ttl time.Duration) (string, error) {
	for i := 0; i < 10; i++ {
		code, err := generateCode()
		if err != nil {
			return "", err
		}
		now := time.Now()
		s.mu.Lock()
		if existing, exists := s.entries[code]; exists && existing.expiresAt.After(now) {
			s.mu.Unlock()
			continue
		}
		s.entries[code] = &pairEntry{telegramID: telegramID, expiresAt: now.Add(ttl)}
		s.mu.Unlock()
		return code, nil
	}
	return "", errors.New("code generation: too many collisions")
}

func (s *pairStore) lookup(code string) (pairEntry, lookupStatus) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[code]
	if !exists {
		return pairEntry{}, lookupNotFound
	}
	if !entry.expiresAt.After(now) {
		delete(s.entries, code)
		return pairEntry{}, lookupExpired
	}
	return *entry, lookupOK
}

func (s *pairStore) delete(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, code)
}

// recordFailure increments failed attempts for an existing code and burns it when max is reached. Returns true if the code was burned.
func (s *pairStore) recordFailure(code string, max int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[code]
	if !exists {
		return false
	}
	entry.failedAttempts++
	if entry.failedAttempts >= max {
		delete(s.entries, code)
		return true
	}
	return false
}

func (s *pairStore) cleanupExpired() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, entry := range s.entries {
		if !entry.expiresAt.After(now) {
			delete(s.entries, code)
		}
	}
}

func (s *pairStore) runCleanup(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.cleanupExpired()
		case <-stop:
			return
		}
	}
}
