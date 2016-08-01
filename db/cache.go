package db

import (
	"sync"
	"sync/atomic"
)

type cache struct {
	mu sync.Mutex

	growThreshold int32
	mask          uint32
	buckets       []*bucket

	writes int32
	nodes  int32
}

func newCache() *cache {
	c := cache{}

	c.buckets = make([]*bucket, threshold)
	c.mask = threshold - 1
	c.growThreshold = int32(threshold * threshold)

	for k := range c.buckets {
		c.buckets[k] = &bucket{}
	}

	return &c
}

func (c *cache) delete(db *DB, key string) error {
	atomic.AddInt32(&c.writes, 1)
	defer func() { atomic.AddInt32(&c.writes, -1) }()

	h := hash([]byte(key), seed)

	k := h & c.mask

	b := c.buckets[k]

	b.delete(key)

	return nil
}

func (c *cache) keys() []string {
	var keys []string
	for _, b := range c.buckets {
		keys = append(keys, b.keys()...)
	}

	return keys
}

func (c *cache) write(db *DB, key string, val []byte, ttl *int) error {
	atomic.AddInt32(&c.writes, 1)
	defer func() { atomic.AddInt32(&c.writes, -1) }()

	h := hash([]byte(key), seed)

	k := h & c.mask

	b := c.buckets[k]

	b.save(key, h, val, ttl)

	if grow := atomic.AddInt32(&c.nodes, 1) >= c.growThreshold; grow {
		db.once.Do(func() {
			go db.grow()
		})
	}

	return nil
}

func (c *cache) read(key string) ([]byte, error) {
	k := hash([]byte(key), seed)

	h := k & c.mask

	b := c.buckets[h]

	if b == nil {
		return nil, nil
	}

	if n := b.lookup(key); n != nil {
		return n.Value, nil
	}

	return nil, nil
}
