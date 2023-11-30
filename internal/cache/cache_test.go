package lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	checkNotInCache := func(val interface{}, ok bool) {
		require.False(t, ok)
		require.Nil(t, val)
	}

	checkIsInCache := func(expected interface{}, val interface{}, ok bool) {
		require.True(t, ok)
		require.Equal(t, expected, val)
	}

	t.Run(
		"empty cache", func(t *testing.T) {
			c := New(10, time.Minute)

			val, ok := c.Get("aaa")
			checkNotInCache(val, ok)

			val, ok = c.Get("bbb")
			checkNotInCache(val, ok)
		},
	)

	t.Run(
		"simple", func(t *testing.T) {
			c := New(5, time.Minute)

			wasInCache := c.Set("aaa", 100)
			require.False(t, wasInCache)

			wasInCache = c.Set("bbb", 200)
			require.False(t, wasInCache)

			val, ok := c.Get("aaa")
			checkIsInCache(100, val, ok)

			val, ok = c.Get("bbb")
			checkIsInCache(200, val, ok)

			wasInCache = c.Set("aaa", 300)
			require.True(t, wasInCache)

			val, ok = c.Get("aaa")
			checkIsInCache(300, val, ok)

			val, ok = c.Get("ccc")
			checkNotInCache(val, ok)
		},
	)

	t.Run(
		"purge logic", func(t *testing.T) {
			c := New(3, time.Minute)
			c.Set("a", 100)       // [a]
			c.Set("b", 200)       // [b, a]
			c.Set("c", 300)       // [c, b, a]
			c.Set("d", 400)       // [d, c, b]
			val, ok := c.Get("a") // [d, c, b]
			checkNotInCache(val, ok)

			wasInCache := c.Set("a", 500) // [a, d, c]
			require.False(t, wasInCache)

			val, ok = c.Get("b") // [a, d, c]
			checkNotInCache(val, ok)

			val, ok = c.Get("c") // [c, a, d]
			checkIsInCache(300, val, ok)

			val, ok = c.Get("d") // [d, c, a]
			checkIsInCache(400, val, ok)

			wasInCache = c.Set("a", 600) // [a, d, c]
			require.True(t, wasInCache)

			val, ok = c.Get("c") // [c, a, d]
			checkIsInCache(300, val, ok)

			wasInCache = c.Set("b", 700) // [b, c, a]
			require.False(t, wasInCache)

			val, ok = c.Get("d")
			checkNotInCache(val, ok)
		},
	)

	t.Run(
		"clear cache", func(t *testing.T) {
			c := New(3, time.Minute)

			c.Set("aaa", 100)
			c.Set("bbb", 200)
			c.Set("ccc", 300)

			val, ok := c.Get("aaa")
			checkIsInCache(100, val, ok)

			val, ok = c.Get("bbb")
			checkIsInCache(200, val, ok)

			val, ok = c.Get("ccc")
			checkIsInCache(300, val, ok)

			c.Clear()

			val, ok = c.Get("aaa")
			checkNotInCache(val, ok)

			val, ok = c.Get("bbb")
			checkNotInCache(val, ok)

			val, ok = c.Get("ccc")
			checkNotInCache(val, ok)

			wasInCache := c.Set("aaa", 400)
			require.False(t, wasInCache)

			val, ok = c.Get("aaa")
			checkIsInCache(400, val, ok)
		},
	)
}

func TestCacheMultithreading(t *testing.T) { //nolint: revive
	c := New(10, time.Minute)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000)))) //nolint: gosec
		}
	}()

	wg.Wait()
}

func TestCacheTTL(t *testing.T) {
	c := New(10, time.Millisecond)
	c.Set("aaa", 100)

	val, ok := c.Get("aaa")
	require.True(t, ok)
	require.Equal(t, 100, val)

	time.Sleep(time.Millisecond)

	val, ok = c.Get("aaa")
	require.False(t, ok)
	require.Nil(t, val)
}
