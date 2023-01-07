package internal

import "errors"

var ErrNoMoreElems error = errors.New("no more elements in the stack")
var ErrOutOfBounds error = errors.New("new offset is out of the cache bounds")

// Cache is a simple collection with useful methods
type Cache[T any] struct {
	arr    *[]T
	size   int
	offset int
}

// NewCache creates a new Cache with a capacity of 20, enought for many parsers
func NewCache[T any]() *Cache[T] {
	arr := make([]T, 0, 20)
	return &Cache[T]{
		arr:    &arr,
		offset: -1,
	}
}

// Push pushes an element onto the Cache
func (s *Cache[T]) Push(elem T) {
	(*s.arr) = append((*s.arr), elem)
	s.size++
	s.offset = s.size - 1
}

// Pos returns the current offset of the cache
func (s *Cache[T]) Pos() int {
	return s.offset
}

// SetOffset sets the current offset of the cache.
// Returns an error if offset specify an index out of the bounds.
func (s *Cache[T]) SetOffset(offset int) error {
	if offset >= s.size {
		return ErrOutOfBounds
	}

	s.offset = offset
	return nil
}

// Back subtract from the current offset n times.
// If the resulting offset is less than 0, its set to 0.
func (s *Cache[T]) Back(n int) {
	s.offset -= n
	if s.offset < 0 {
		s.offset = 0
	}
}

// Current returns the element at the current cache's offset
func (s *Cache[T]) Current() *T {
	elem := (*s.arr)[s.offset]
	return &elem
}

// NextHas returns true if s.offset+1 points to a valid element
func (s *Cache[T]) NextHas() bool {
	return s.offset+1 < s.size
}

// Advance advances one position and returns the element at the position
func (s *Cache[T]) Advance() *T {
	if s.offset+1 < s.size {
		s.offset++
	}

	curr := s.Current()
	if curr != nil {
		return curr
	}

	return nil
}

// Before returns the element before element at s.offset
func (s *Cache[T]) Before() *T {
	if s.offset == 0 {
		return nil
	}

	offset := s.offset - 1
	elem := (*s.arr)[offset]
	return &elem
}
