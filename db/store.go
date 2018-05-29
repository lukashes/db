package db

import (
	"sync/atomic"
)

type store struct {
	growThreshold int32
	mask          uint32
	buckets       []*bucket

	// Atomic counters
	writes int32
	nodes  int32
}

func newStore() *store {
	c := store{}

	c.buckets = make([]*bucket, growingSize)
	c.mask = growingSize - 1
	c.growThreshold = int32(growingSize * growingSize)

	for k := range c.buckets {
		c.buckets[k] = new(bucket)
	}

	return &c
}

func (c *store) delete(db *DB, key string) error {
	atomic.AddInt32(&c.writes, 1)
	defer func() { atomic.AddInt32(&c.writes, -1) }()

	h := hash([]byte(key), seed)
	k := h & c.mask
	b := c.buckets[k]

	b.delete(key)

	return nil
}

func (c *store) keys() []string {
	var keys []string
	for _, b := range c.buckets {
		keys = append(keys, b.keys()...)
	}

	return keys
}

func (c *store) write(db *DB, key string, val interface{}, ttl *int) error {
	atomic.AddInt32(&c.writes, 1)

	h := hash([]byte(key), seed)
	k := h & c.mask
	b := c.buckets[k]

	b.save(key, h, val, ttl)

	if grow := atomic.AddInt32(&c.nodes, 1) >= c.growThreshold; grow {
		db.once.Do(func() {
			go db.grow()
		})
	}

	atomic.AddInt32(&c.writes, -1)
	return nil
}

func (c *store) read(key string) (*node, error) {
	k := hash([]byte(key), seed)
	h := k & c.mask
	b := c.buckets[h]

	if b == nil {
		return nil, ErrNotFound
	}

	if n := b.lookup(key); n != nil {
		return n, nil
	}

	return nil, ErrNotFound
}
