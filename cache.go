package discache

import (
	"container/list"
	"log"
	"sync"

	"golang.org/x/sync/singleflight"
)

var db = map[string][]byte{
	"a": []byte("Hello from a"),
	"b": []byte("Hello from b"),
	"c": []byte("Hello from c"),
}

type entry struct {
	key   string
	value []byte
}

func (e *entry) Size() int64 {
	return int64(len(e.value)) + int64(len(e.key))
}

type Cache struct {
	mut          sync.Mutex
	singleFlight singleflight.Group

	maxByte int64
	curByte int64
	data    *list.List
	cache   map[string]*list.Element

	onEvicted func(key string, value []byte)
}

func NewCache(maxByte int64, onEvicted func(key string, value []byte)) *Cache {
	return &Cache{
		maxByte:   maxByte,
		curByte:   0,
		data:      list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mut.Lock()
	defer c.mut.Unlock()
	if e := c.cache[key]; e != nil {
		c.data.MoveToBack(e)
		log.Printf("Get from cache %s", key)
		return e.Value.(*entry).value, true
	} else {
		v, ok, _ := c.singleFlight.Do(key, func() (any, error) {
			return c.getFromDB(key)
		})
		if ok != nil {
			log.Printf("Key %s not found in db", key)
			return nil, false
		}
		log.Printf("Get from db %s", key)
		c.add(key, v.([]byte))
		return v.([]byte), true
	}
}

func (c *Cache) add(key string, value []byte) {
	c.mut.Lock()
	defer c.mut.Unlock()
	if ele, ok := c.cache[key]; ok {
		c.curByte -= int64(len(ele.Value.(*entry).value))
		c.data.MoveToBack(ele)
		ele.Value = value
	} else {
		e := c.data.PushBack(&entry{key: key, value: value})
		c.cache[key] = e
		c.curByte += int64(len(key))
	}
	c.curByte += int64(len(value))
	for c.curByte > c.maxByte {
		c.evict()
	}
}

func (c *Cache) remove(key string) bool {
	c.mut.Lock()
	defer c.mut.Unlock()
	if e := c.cache[key]; e != nil {
		c.data.Remove(e)
		delete(c.cache, key)
		c.curByte -= int64(len(e.Value.(*entry).value))
		return true
	}
	return false
}

func (c *Cache) Len() int {
	return c.data.Len()
}

func (c *Cache) Size() int64 {
	return c.curByte
}

func (c *Cache) evict() {
	kv := c.data.Front().Value.(*entry)
	c.remove(kv.key)
	if c.onEvicted != nil {
		c.onEvicted(kv.key, kv.value)
	}
}

func (c *Cache) getFromDB(key string) ([]byte, error) {
	if v, ok := db[key]; ok {
		return v, nil
	}
	return nil, nil
}
