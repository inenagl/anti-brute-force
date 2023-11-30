package bucketstorage

import (
	"sync"
	"time"

	"github.com/inenagl/anti-brute-force/internal/ratelimit"
)

type Storage struct {
	mu   sync.Mutex
	data map[interface{}]ratelimit.Bucket
	ttl  time.Duration
}

func New(ttl time.Duration) *Storage {
	return &Storage{
		mu:   sync.Mutex{},
		ttl:  ttl,
		data: make(map[interface{}]ratelimit.Bucket),
	}
}

func (s *Storage) Set(key string, bucket ratelimit.Bucket) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = bucket
}

func (s *Storage) Get(key string) (ratelimit.Bucket, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, ok := s.data[key]
	return bucket, ok
}

func (s *Storage) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
}

func (s *Storage) ClearByTTL() {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := time.Now().Add(-1 * s.ttl)
	for k, v := range s.data {
		if v.GetLastTS().Before(t) {
			delete(s.data, k)
		}
	}
}

func (s *Storage) ClearAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[interface{}]ratelimit.Bucket)
}
