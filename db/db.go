package db

import (
	"errors"
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	errEmptyKey = errors.New("empty key")
)

const (
	threshold = 1 << 4
	seed      = 125
)

type DB struct {
	mu   sync.Mutex
	once sync.Once

	h unsafe.Pointer
	t unsafe.Pointer
}

func New() *DB {
	db := new(DB)

	db.h = unsafe.Pointer(newCache())

	return db
}

func (db *DB) grow() {
	db.mu.Lock()
	defer func() {
		db.once = sync.Once{}
		db.mu.Unlock()
	}()

	s := db.head()

	buckets := len(s.buckets) << 1
	d := &cache{
		buckets:       make([]*bucket, buckets),
		mask:          uint32(buckets) - 1,
		growThreshold: int32(buckets * threshold),
	}

	for k := range d.buckets {
		d.buckets[k] = new(bucket)
	}

	db.t = db.h
	atomic.StorePointer(&db.h, unsafe.Pointer(d))

	for {
		if cnt := atomic.LoadInt32(&db.tail().writes); cnt == 0 {
			break
		}
	}

	for k := range d.buckets {
		d.buckets[k].do(func(b *bucket) {
			exists := map[string]bool{}
			for _, n := range b.Nodes {
				exists[n.Key] = true
			}
			pb := s.buckets[uint32(k)&s.mask]

			for _, n := range pb.Nodes {
				if _, ok := exists[n.Key]; !ok && n.Hash&d.mask == uint32(k) && n.isAlive() {
					b.Nodes = append(b.Nodes, n)
					atomic.AddInt32(&d.nodes, 1)
				}
			}
		})
	}

	atomic.StorePointer(&db.t, nil)
}

func (db *DB) head() *cache {
	c := atomic.LoadPointer(&db.h)
	return (*cache)(c)
}

func (db *DB) tail() *cache {
	c := atomic.LoadPointer(&db.t)
	return (*cache)(c)
}

// Delete marks keys as deleted
func (db *DB) Delete(key string) error {

	if len(key) == 0 {
		return errEmptyKey
	}

	return db.head().delete(db, key)
}

// Keys returns slice of available keys
// It goes for all buckets one by one
// and get keys.
func (db *DB) Keys() []string {
	return db.head().keys()
}

// Write sets new value or rewrite exists
func (db *DB) Write(key string, val []byte, ttl *int) error {

	if len(key) == 0 {
		return errEmptyKey
	}

	return db.head().write(db, key, val, ttl)
}

// Read returns value associated with key or nil
func (db *DB) Read(key string) ([]byte, error) {

	res, err := db.head().read(key)

	if err != nil {
		return res, err
	}

	if res == nil && db.tail() != nil {
		res, err = db.tail().read(key)
	}

	return res, err
}

// Exists checking key existing
func (db *DB) Exists(key string) (bool, error) {

	val, err := db.head().read(key)

	return val != nil, err
}
