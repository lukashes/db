package db

import (
	"sync"
	"time"
)

type bucket struct {
	mu sync.RWMutex

	// Pointer to first item at the bucket
	nodes *node
}

func (b *bucket) do(c func(b *bucket)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	c(b)
}

// Soft delete
//
// Set node as expired
func (b *bucket) delete(key string) {
	b.mu.Lock()

	node, found := b.find(key)

	if found {
		node.exp = -1
	}

	b.mu.Unlock()
}

func (b *bucket) keys() []string {
	b.mu.RLock()

	keys := make([]string, 0, 20)
	for n := b.nodes; n != nil; n = n.next {
		if n.isAlive() {
			keys = append(keys, n.key)
		}
	}

	b.mu.RUnlock()

	return keys
}

// Finds node with provided key or last node in the chain
func (b *bucket) find(key string) (*node, bool) {
	if b.nodes == nil {
		return nil, false
	}

	var n *node
	for n = b.nodes; n.next != nil; n = n.next {
		if n.key == key {
			return n, true
		}
	}

	if n != nil && n.key == key {
		return n, true
	}

	return n, false
}

func (b *bucket) save(key string, hash uint32, val interface{}, ttl *int) error {
	b.mu.Lock()

	n, found := b.find(key)

	exp := 0
	if ttl != nil { // todo: do not generate time every call
		exp = time.Now().Second() + *ttl
	}

	if !found {
		if n == nil {
			n = &node{
				key:  key,
				exp:  exp,
				hash: hash,
			}
			b.nodes = n
		} else {
			n.next = &node{
				key:  key,
				exp:  exp,
				hash: hash,
			}
			n = n.next
		}
	}

	switch t := val.(type) {
	case []byte:
		n.value = t
		n.tipe = TypeHash
	case []string:
		n.list = t
		n.tipe = TypeList
	case map[string]string:
		n.dict = t
		n.tipe = TypeDict
	default:
		b.mu.Unlock()
		return ErrInvalidType
	}

	b.mu.Unlock()
	return nil
}

func (b *bucket) lookup(key string) *node {
	b.mu.RLock()

	n, found := b.find(key)
	if !found {
		b.mu.RUnlock()
		return nil
	}

	if !n.isAlive() {
		b.mu.RUnlock()
		return nil
	}

	b.mu.RUnlock()
	return n
}
