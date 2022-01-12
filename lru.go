package lru

import (
	"sync"

	"github.com/optakt/golang-lru/simplelru"
)

const (
	// DefaultEvictedBufferSize defines the default buffer size to store evicted key/val
	DefaultEvictedBufferSize = 16
)

// Cache is a thread-safe fixed size LRU cache.
type Cache struct {
	lru         *simplelru.LRU
	evictedKeys []interface{}
	evictedVals []interface{}
	lock        sync.RWMutex
}

// New creates an LRU of the given size.
func New(size int) (*Cache, error) {
	return NewWithEvict(size, nil)
}

// NewWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewWithEvict(size int, onEvicted func(k, v interface{}, used int) bool) (*Cache, error) {
	// create a cache with default settings
	c := &Cache{}
	if onEvicted != nil {
		c.initEvictBuffers()
	}

	var err error
	c.lru, err = simplelru.NewLRU(size, onEvicted)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Cache) initEvictBuffers() {
	c.evictedKeys = make([]interface{}, 0, DefaultEvictedBufferSize)
	c.evictedVals = make([]interface{}, 0, DefaultEvictedBufferSize)
}

// onEvicted saves evicted keys and values.
func (c *Cache) onEvicted(k, v interface{}) bool {
	c.evictedKeys = append(c.evictedKeys, k)
	c.evictedVals = append(c.evictedVals, v)
	return true
}

// Purge is used to completely clear the cache.
func (c *Cache) Purge() {
	c.lock.Lock()
	c.lru.Purge()
	if len(c.evictedKeys) > 0 {
		c.initEvictBuffers()
	}
	c.lock.Unlock()
}

// Add adds a value to the cache. Returns true if an eviction occurred.
func (c *Cache) Add(key, value interface{}) (evicted bool) {
	c.lock.Lock()
	evicted = c.lru.Add(key, value)
	if evicted {
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.lock.Unlock()
	return
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key interface{}) (value interface{}, ok bool) {
	c.lock.Lock()
	value, ok = c.lru.Get(key)
	c.lock.Unlock()
	return value, ok
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *Cache) Contains(key interface{}) bool {
	c.lock.RLock()
	containKey := c.lru.Contains(key)
	c.lock.RUnlock()
	return containKey
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *Cache) Peek(key interface{}) (value interface{}, ok bool) {
	c.lock.RLock()
	value, ok = c.lru.Peek(key)
	c.lock.RUnlock()
	return value, ok
}

// ContainsOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *Cache) ContainsOrAdd(key, value interface{}) (ok, evicted bool) {
	c.lock.Lock()
	if c.lru.Contains(key) {
		c.lock.Unlock()
		return true, false
	}
	evicted = c.lru.Add(key, value)
	if evicted {
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.lock.Unlock()
	return false, evicted
}

// PeekOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *Cache) PeekOrAdd(key, value interface{}) (previous interface{}, ok, evicted bool) {
	c.lock.Lock()
	previous, ok = c.lru.Peek(key)
	if ok {
		c.lock.Unlock()
		return previous, true, false
	}
	evicted = c.lru.Add(key, value)
	if evicted {
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.lock.Unlock()
	return nil, false, evicted
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key interface{}) (present bool) {
	c.lock.Lock()
	present = c.lru.Remove(key)
	if present {
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.lock.Unlock()
	return
}

// Resize changes the cache size.
func (c *Cache) Resize(size int) (evicted int) {
	c.lock.Lock()
	evicted = c.lru.Resize(size)
	if evicted > 0 {
		c.initEvictBuffers()
	}
	c.lock.Unlock()
	return evicted
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() (key, value interface{}, ok bool) {
	c.lock.Lock()
	key, value, ok = c.lru.RemoveOldest()
	if ok {
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.lock.Unlock()
	return
}

// GetOldest returns the oldest entry
func (c *Cache) GetOldest() (key, value interface{}, ok bool) {
	c.lock.RLock()
	key, value, ok = c.lru.GetOldest()
	c.lock.RUnlock()
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *Cache) Keys() []interface{} {
	c.lock.RLock()
	keys := c.lru.Keys()
	c.lock.RUnlock()
	return keys
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	c.lock.RLock()
	length := c.lru.Len()
	c.lock.RUnlock()
	return length
}
