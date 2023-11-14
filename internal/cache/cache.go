package lrucache

import (
	"sync"
	"time"
)

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	mu       sync.Mutex
	capacity int
	queue    List
	items    map[Key]*ListItem
	ttl      time.Duration
}

type cacheItemValue struct {
	k Key
	v interface{}
	t time.Time
}

func New(capacity int, ttl time.Duration) Cache {
	return &lruCache{
		mu:       sync.Mutex{},
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
		ttl:      ttl,
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]

	civ := cacheItemValue{k: key, v: value, t: time.Now()}
	if ok {
		item.Value = civ
		c.queue.MoveToFront(item)
	} else {
		item = c.queue.PushFront(civ)
		c.items[key] = item
		if c.queue.Len() > c.capacity {
			toDelete := c.queue.Back()
			c.queue.Remove(toDelete)
			delete(c.items, toDelete.Value.(cacheItemValue).k)
		}
	}

	return ok
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result interface{}

	item, ok := c.items[key]
	if ok {
		if time.Since(item.Value.(cacheItemValue).t) >= c.ttl {
			c.queue.Remove(item)
			delete(c.items, key)
			return result, false
		}

		result = item.Value.(cacheItemValue).v
		c.queue.MoveToFront(item)
	}

	return result, ok
}

func (c *lruCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}
