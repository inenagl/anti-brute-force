package bucketstorage

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/inenagl/anti-brute-force/internal/ratelimit"
	"github.com/jackc/fake"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	rand.NewSource(time.Now().UnixNano())

	s := New(time.Minute)

	bucket := ratelimit.NewBucket(rand.Intn(100), rand.Float64()) //nolint:gosec
	key := fake.Word()

	s.Set(key, bucket)
	require.Equal(t, bucket, s.data[key])
}

func TestGet(t *testing.T) {
	rand.NewSource(time.Now().UnixNano())

	s := New(time.Minute)

	bucket := ratelimit.NewBucket(rand.Intn(100), rand.Float64()) //nolint:gosec
	key := fake.Word()
	var key2 string
	for key2 = fake.Word(); key2 == key; {
		key2 = fake.Word()
	}

	s.data[key] = bucket

	res, ok := s.Get(key)
	require.True(t, ok)
	require.Equal(t, bucket, res)

	res, ok = s.Get(key2)
	require.False(t, ok)
	require.Nil(t, res)
}

func TestRemove(t *testing.T) {
	rand.NewSource(time.Now().UnixNano())

	s := New(time.Minute)

	bucket := ratelimit.NewBucket(rand.Intn(100), rand.Float64()) //nolint:gosec
	key := fake.Word()

	s.data[key] = bucket

	s.Remove(key)
	require.Equal(t, 0, len(s.data))
}

func TestClearAll(t *testing.T) {
	rand.NewSource(time.Now().UnixNano())

	s := New(time.Minute)

	var bucket ratelimit.Bucket
	var key string
	for i := 0; i < 10; i++ {
		bucket = ratelimit.NewBucket(rand.Intn(100), rand.Float64()) //nolint:gosec
		key = fake.Word()
		s.data[key] = bucket
	}

	s.ClearAll()
	require.Equal(t, 0, len(s.data))
}

func TestClearByTTL(t *testing.T) {
	rand.NewSource(time.Now().UnixNano())

	ttl := 100 * time.Millisecond
	s := New(ttl)

	var bucket ratelimit.Bucket
	var key string
	for i := 0; i < 10; i++ {
		bucket = ratelimit.NewBucket(rand.Intn(100), rand.Float64()) //nolint:gosec
		key = strconv.Itoa(i)
		s.data[key] = bucket
	}
	s.ClearByTTL()
	require.Equal(t, 10, len(s.data))

	time.Sleep(ttl)
	i := 0
	for _, v := range s.data {
		v.Increment()
		i++
		if i == 2 {
			break
		}
	}
	s.ClearByTTL()
	require.Equal(t, 2, len(s.data))
}
