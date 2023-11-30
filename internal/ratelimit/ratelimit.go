package ratelimit

import (
	"math"
	"sync"
	"time"
)

type Bucket interface {
	Increment() bool
	GetLastTS() time.Time
	GetCount() int
	GetRate() float64
}

type bucket struct {
	mu     sync.Mutex
	count  int       // current count of units
	cap    int       // capacity of bucket in units
	rate   float64   // rate of leaking in units per second
	lastTS time.Time // last increment attempt time
}

// Increment Реализация алгоритма leaky bucket.
func (b *bucket) Increment() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	// Вычисляем сколько должно быть единиц в бакете с учетом скорости изливания
	// и времени прошедшего с последнего пополнения бакета,
	// и инкрементим это число.
	b.count = int(math.Ceil(float64(b.count)-now.Sub(b.lastTS).Seconds()*b.rate)) + 1

	// Если долго не было инкрементов, то расчётное значение может уйти в минус,
	// такого нельзя допускать.
	if b.count <= 0 {
		b.count = 1
	}

	// Выставляем текущее время пополнения бакета.
	b.lastTS = now

	// Если бакет переполнен, устанавливаем счетчик в максимально возможное значение и возвращаем неудачу.
	if b.count > b.cap {
		b.count = b.cap
		return false
	}

	return true
}

func (b *bucket) GetLastTS() time.Time {
	return b.lastTS
}

func (b *bucket) GetCount() int {
	return b.count
}

func (b *bucket) GetRate() float64 {
	return b.rate
}

func NewBucket(capacity int, rate float64) Bucket {
	return &bucket{
		mu:     sync.Mutex{},
		count:  0,
		cap:    capacity,
		rate:   rate,
		lastTS: time.Now(),
	}
}
