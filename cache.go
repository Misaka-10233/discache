package discache

import (
	"container/list"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/singleflight"
)

type entry struct {
	key   string
	value []byte
}

func (e *entry) Size() int64 {
	return int64(len(e.value)) + int64(len(e.key))
}

type Cache struct {
	rwMut        sync.RWMutex
	singleFlight singleflight.Group

	maxByte int64
	curByte int64
	data    *list.List
	cache   map[string]*list.Element

	requestCount int32
	hitCount     int32

	onEvicted func(key string, value []byte)
	onMissed  func(key string) ([]byte, error)
}

func NewCache(maxByte int64, onEvicted func(key string, value []byte), onMissed func(key string) ([]byte, error)) *Cache {
	return &Cache{
		rwMut:        sync.RWMutex{},
		singleFlight: singleflight.Group{},
		maxByte:      maxByte,
		curByte:      0,
		data:         list.New(),
		cache:        make(map[string]*list.Element),
		onEvicted:    onEvicted,
		onMissed:     onMissed,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.countRequest(1)
	c.rwMut.RLock()
	e := c.cache[key]
	c.rwMut.RUnlock()
	if e != nil {
		c.rwMut.Lock()
		c.data.MoveToBack(e)
		c.rwMut.Unlock()
		log.Printf("Get from cache %s", key)
		c.countHit(1)
		return e.Value.(*entry).value, true
	} else {
		v, err, _ := c.singleFlight.Do(key, func() (any, error) {
			return c.onMissed(key)
		})
		if err != nil || v == nil {
			log.Printf("Key %s not found in db, err: %v", key, err)
			return nil, false
		}
		log.Printf("Get from db %s", key)
		c.Add(key, v.([]byte))
		return v.([]byte), true
	}
}

func (c *Cache) Add(key string, value []byte) {
	c.rwMut.Lock()
	if e := c.cache[key]; e != nil {
		c.curByte -= e.Value.(*entry).Size()
		e.Value = &entry{key: key, value: value}
		c.data.MoveToBack(e)
		c.curByte += e.Value.(*entry).Size()
		c.rwMut.Unlock()
		return
	}
	e := c.data.PushBack(&entry{key: key, value: value})
	c.cache[key] = e
	c.curByte += e.Value.(*entry).Size()
	c.rwMut.Unlock()
	for c.curByte > c.maxByte {
		c.evict()
	}
}

func (c *Cache) remove(key string) error {
	c.rwMut.Lock()
	defer c.rwMut.Unlock()
	if e := c.cache[key]; e != nil {
		c.data.Remove(e)
		delete(c.cache, key)
		c.curByte -= e.Value.(*entry).Size()
		return nil
	}
	return fmt.Errorf("key %s not found in cache", key)
}

func (c *Cache) Len() int {
	c.rwMut.RLock()
	defer c.rwMut.RUnlock()
	return c.data.Len()
}

func (c *Cache) Size() int64 {
	c.rwMut.RLock()
	defer c.rwMut.RUnlock()
	return c.curByte
}

func (c *Cache) evict() {
	c.rwMut.RLock()
	kv := c.data.Front().Value.(*entry)
	c.rwMut.RUnlock()
	_ = c.remove(kv.key)
	if c.onEvicted != nil {
		c.onEvicted(kv.key, kv.value)
	}
}

func (c *Cache) countRequest(delta int32) {
	atomic.AddInt32(&c.requestCount, delta)
}

func (c *Cache) countHit(delta int32) {
	atomic.AddInt32(&c.hitCount, delta)
}

func (c *Cache) GetRequestCount() int32 {
	return atomic.LoadInt32(&c.requestCount)
}
func (c *Cache) GetHitCount() int32 {
	return atomic.LoadInt32(&c.hitCount)
}
