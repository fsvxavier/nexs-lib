package cache

import (
	"container/list"
	"sync"
)

// LRUCache implements a thread-safe LRU cache
type LRUCache struct {
	capacity int
	list     *list.List
	items    map[interface{}]*list.Element
	lock     sync.RWMutex
}

type entry struct {
	key   interface{}
	value interface{}
}

// NewLRUCache creates a new LRU cache with the given capacity
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		list:     list.New(),
		items:    make(map[interface{}]*list.Element),
	}
}

// Get gets a value from the cache
func (c *LRUCache) Get(key interface{}) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if element, exists := c.items[key]; exists {
		c.list.MoveToFront(element)
		return element.Value.(*entry).value, true
	}
	return nil, false
}

// Set sets a value in the cache
func (c *LRUCache) Set(key interface{}, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if element, exists := c.items[key]; exists {
		c.list.MoveToFront(element)
		element.Value.(*entry).value = value
		return
	}

	if c.list.Len() >= c.capacity {
		c.removeLRU()
	}

	element := c.list.PushFront(&entry{key, value})
	c.items[key] = element
}

// Remove removes a key from the cache
func (c *LRUCache) Remove(key interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if element, exists := c.items[key]; exists {
		c.removeElement(element)
	}
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.list = list.New()
	c.items = make(map[interface{}]*list.Element)
}

// Len returns the number of items in the cache
func (c *LRUCache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.list.Len()
}

func (c *LRUCache) removeLRU() {
	element := c.list.Back()
	if element != nil {
		c.removeElement(element)
	}
}

func (c *LRUCache) removeElement(element *list.Element) {
	c.list.Remove(element)
	delete(c.items, element.Value.(*entry).key)
}
