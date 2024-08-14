package set

import (
	"sync"

	"golang.org/x/exp/maps"
)

// Set can be used to check if a given key exists in a set or not. It uses a map with an empty struct as values to
// prevent extra memory allocations. Set is thread-safe.
type Set[T comparable] struct {
	mu       sync.RWMutex
	elements map[T]struct{}
}

// Add can be used to add a given key to the set. Returns true of key is new. If key is duplicate returns false.
func (s *Set[T]) Add(element T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.elements[element]; ok {
		return false
	}

	s.elements[element] = struct{}{}

	return true
}

// Remove deletes a given key from the set. Returns true if key is removed. If key does not exist in the set, Returns false.
func (s *Set[T]) Remove(element T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.elements[element]; !ok {
		return false
	}

	delete(s.elements, element)

	return true
}

// Contains checks whether a given key exists in the set or not.
func (s *Set[T]) Contains(element T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.elements[element]

	return exists
}

// ToSimpleMap converts the set to the standard golang map.
func (s *Set[T]) ToSimpleMap() map[T]struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return maps.Clone(s.elements)
}

func New[T comparable]() *Set[T] {
	return &Set[T]{
		elements: make(map[T]struct{}),
	}
}
