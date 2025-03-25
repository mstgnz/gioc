package gioc

import "sync"

// Scope represents the lifetime of a component in the IoC container
type Scope int

// ScopeID represents a unique identifier for a scope
type ScopeID string

// ScopeContext maintains instances within a specific scope
type ScopeContext struct {
	id        ScopeID
	instances map[uintptr]any
	mu        sync.RWMutex
}

// Get returns an instance from the scope context
func (s *ScopeContext) Get(key uintptr) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	instance, exists := s.instances[key]
	return instance, exists
}

// Set stores an instance in the scope context
func (s *ScopeContext) Set(key uintptr, instance any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[key] = instance
}

// Cleanup removes all instances from the scope context
func (s *ScopeContext) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Create a new map to avoid any race conditions with existing references
	s.instances = make(map[uintptr]any)
}
