package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	cache := NewCache[int]()
	cache.Push(5)
	cache.Push(3)

	require.Equal(t, 2, cache.size)
	require.Equal(t, 1, cache.offset)

	elem := cache.Current()
	require.Equal(t, 3, *elem)

	cache.SetOffset(0)
	elem = cache.Current()
	require.Equal(t, 5, *elem)

	err := cache.SetOffset(2)
	require.ErrorIs(t, err, ErrOutOfBounds)

	cache.Push(25)
	elem = cache.Current()
	require.Equal(t, 25, *elem)
}
