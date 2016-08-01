package db

import (
	"sync"
	"time"
)

type node struct {
	Key        string
	Expiration int
	Value      []byte
	Hash       uint32
}

func (n *node) isAlive() bool {
	if n == nil || n.Value == nil {
		return false
	}

	if n.Expiration == 0 || n.Expiration > time.Now().Second() {
		return true
	}

	return false
}

type bucket struct {
	mu sync.RWMutex

	Nodes []node

	locked bool
}

func (b *bucket) do(c func(b *bucket)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	c(b)
}

func (b *bucket) delete(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	k := b.find(key)

	if k > -1 {
		b.Nodes[k].Expiration = -1
	}
}

func (b *bucket) keys() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	keys := make([]string, 0, len(b.Nodes))
	for k := range b.Nodes {
		if b.Nodes[k].isAlive() {
			keys = append(keys, b.Nodes[k].Key)
		}
	}

	return keys
}

func (b *bucket) find(key string) (pos int) {
	for k, n := range b.Nodes {
		if n.Key == key {
			return k
		}
	}

	return -1
}

func (b *bucket) save(key string, hash uint32, val []byte, ttl *int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	k := b.find(key)

	exp := 0
	if ttl != nil {
		exp = time.Now().Second() + *ttl
	}

	if k < 0 {
		n := node{
			Value:      val,
			Key:        key,
			Expiration: exp,
			Hash:       hash,
		}
		b.Nodes = append(b.Nodes, n)
	} else {
		b.Nodes[k].Expiration = exp
		b.Nodes[k].Value = val
	}

	return nil
}

func (b *bucket) lookup(key string) *node {
	b.mu.RLock()
	defer b.mu.RUnlock()

	k := b.find(key)

	if k < 0 {
		return nil
	}

	if !b.Nodes[k].isAlive() {
		return nil
	}

	n := b.Nodes[k]

	return &n
}
