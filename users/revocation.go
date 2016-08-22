package users

import "time"

// RevocationStore is the interface for a store which
// keeps track of revoked sessions. By default it uses
// an in-memory store
type RevocationStore interface {
	Revoke(id string, trackFor time.Duration)
	IsRevoked(id string) bool
}

// InMemoryRevocationStore keeps track of revoked tokens
// in memory and periodically flushes old tokens
type InMemoryRevocationStore struct {
	revoked       map[string]time.Time
	flushInterval time.Duration
	ticker        *time.Ticker
}

// Start the collection job
func (s *InMemoryRevocationStore) Start() {
	if s.ticker != nil {
		return
	}
	now := time.Now()
	s.ticker = time.NewTicker(s.flushInterval)
	for range s.ticker.C {
		for session, expiry := range s.revoked {
			if now.After(expiry) {
				delete(s.revoked, session)
			}
		}
	}
}

// Stop the collection job
func (s *InMemoryRevocationStore) Stop() {
	s.ticker.Stop()
	s.ticker = nil
}

// Revoke implements the interface method
func (s *InMemoryRevocationStore) Revoke(id string, trackFor time.Duration) {
	s.revoked[id] = time.Now().Add(trackFor)
}

// IsRevoked implements the interface method
func (s *InMemoryRevocationStore) IsRevoked(id string) bool {
	_, inStore := s.revoked[id]
	// if in store, assume revoked
	return inStore
}
