package session

import "time"

// InMemoryRevocationStore keeps track of revoked tokens
// in memory and periodically flushes old tokens
type InMemoryRevocationStore struct {
	revoked       map[string]time.Time
	flushInterval time.Duration
	ticker        *time.Ticker
}

// NewInMemoryRevocationStore returns a new in memory revocation store
// which checks for expired tokens every flushInterval
func NewInMemoryRevocationStore(flushInterval time.Duration) *InMemoryRevocationStore {
	return &InMemoryRevocationStore{
		revoked:       map[string]time.Time{},
		flushInterval: flushInterval,
	}
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
