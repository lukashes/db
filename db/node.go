package db

import "time"

type Type int

const (
	TypeHash Type = iota
	TypeList
	TypeDict
)

type node struct {
	// Actual key and hash based on key
	key  string
	hash uint32

	// Different types for reducing allocations
	value []byte
	list  []string
	dict  map[string]string

	// Meta
	exp  int
	tipe Type
	next *node
}

func (n *node) isAlive() bool {
	if n == nil {
		return false
	}

	if n.exp == 0 || n.exp > time.Now().Second() {
		return true
	}

	return false
}
