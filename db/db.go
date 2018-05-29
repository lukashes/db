package db

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	nodesSize   = 8   // todo: add growing after max nodes count inside bucket
	growingSize = 2   // how much to increase buckets count after growing
	seed        = 125 // for hash function
)

type DB struct {
	mu   sync.Mutex
	once sync.Once

	// All writes should be to the head
	// It means all fresh data are here
	h unsafe.Pointer

	// Tail is temporary old data pointer
	// during moving buckets
	t unsafe.Pointer
}

func New() *DB {
	db := new(DB)

	db.h = unsafe.Pointer(newStore())

	return db
}

func (db *DB) grow() {
	//fmt.Printf("grow started!\n")
	db.mu.Lock()
	defer func() {
		db.once = sync.Once{}
		db.mu.Unlock()
	}()

	tempStore := db.head()
	buckets := len(tempStore.buckets) << 1
	newStore := &store{
		buckets:       make([]*bucket, buckets),
		mask:          uint32(buckets) - 1,
		growThreshold: int32(buckets * growingSize),
	}

	for k := range newStore.buckets {
		newStore.buckets[k] = new(bucket)
	}

	// Snapshot of actual data
	db.t = db.h
	atomic.StorePointer(&db.h, unsafe.Pointer(newStore))

	for {
		if cnt := atomic.LoadInt32(&db.tail().writes); cnt == 0 {
			break
		}
	}

	// Start moving data to the new store
	for k := range newStore.buckets {
		newStore.buckets[k].do(func(b *bucket) {
			var (
				hotter = make(map[string]bool)
				tail   *node
				pre    *node
			)
			for n := b.nodes; n != nil; n = n.next {
				hotter[n.key] = true
				tail = n
			}
			sourceBuck := tempStore.buckets[uint32(k)&tempStore.mask]
			head := sourceBuck.nodes
			for head != nil {
				if _, ok := hotter[head.key]; !ok && head.hash&newStore.mask == uint32(k) && head.isAlive() { // Swap node!
					// Move node from old buck to the new
					if tail == nil {
						b.nodes = head
					} else {
						tail.next = head
					}

					// Change actual element to the next
					tail = head
					head = tail.next

					sourceBuck.mu.Lock()
					// Cut chain of already moved element
					tail.next = nil
					// Rewrite chain of old store
					if pre != nil {
						pre.next = head
					} else {
						sourceBuck.nodes = head
					}
					sourceBuck.mu.Unlock()

					atomic.AddInt32(&newStore.nodes, 1)
				} else {
					pre = head
					head = pre.next
				}
			}
		})
	}

	atomic.StorePointer(&db.t, nil)
}

func (db *DB) head() *store {
	c := atomic.LoadPointer(&db.h)
	return (*store)(c)
}

func (db *DB) tail() *store {
	c := atomic.LoadPointer(&db.t)
	return (*store)(c)
}

// Delete marks keys as deleted
func (db *DB) Delete(key string) error {

	if len(key) == 0 {
		return ErrEmptyKey
	}

	return db.head().delete(db, key)
}

// Keys returns slice of available keys
//
// It goes for all buckets one by one
// and get keys.
func (db *DB) Keys() []string {
	return db.head().keys()
}

// Write sets new value or rewrite already exists
func (db *DB) Write(key string, val []byte, ttl *int) error {

	if len(key) == 0 {
		return ErrEmptyKey
	}

	return db.head().write(db, key, val, ttl)
}

// Read returns value associated with key or nil
func (db *DB) Read(key string) ([]byte, error) {

	node, err := db.head().read(key)
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	if node == nil && db.tail() != nil {
		node, err = db.tail().read(key)
	}

	if node == nil {
		return nil, ErrNotFound
	}

	if node.tipe != TypeHash {
		return nil, ErrInvalidType
	}

	return node.value, err
}

// WriteList writes list data type
func (db *DB) WriteList(key string, val []string, ttl *int) error {

	if len(key) == 0 {
		return ErrEmptyKey
	}

	return db.head().write(db, key, val, ttl)
}

// WriteDict writes dict data type
func (db *DB) WriteDict(key string, val map[string]string, ttl *int) error {

	if len(key) == 0 {
		return ErrEmptyKey
	}

	return db.head().write(db, key, val, ttl)
}

// ReadListIndex returns data by list index
//
// If index or key do not exist returns ErrNoFound
func (db *DB) ReadListIndex(key string, idx int) ([]byte, error) {

	node, err := db.head().read(key)
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	if node == nil && db.tail() != nil {
		if node, err = db.tail().read(key); err != nil {
			return nil, err
		}
	}

	if node == nil {
		return nil, ErrNotFound
	}

	if node.tipe != TypeList {
		return nil, ErrInvalidType
	}

	if len(node.list) <= idx {
		return nil, ErrInvalidIndex
	}

	return []byte(node.list[idx]), err
}

// ReadList returns whole list data
func (db *DB) ReadList(key string) ([]string, error) {

	node, err := db.head().read(key)
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	if node == nil && db.tail() != nil {
		if node, err = db.tail().read(key); err != nil {
			return nil, err
		}
	}

	if node == nil {
		return nil, ErrNotFound
	}

	if node.tipe != TypeList {
		return nil, ErrInvalidType
	}

	return node.list, err
}

// ReadDictIndex returns data by dict index
//
// If index or key do not exist returns ErrNoFound
func (db *DB) ReadDictIndex(key string, idx string) ([]byte, error) {

	node, err := db.head().read(key)
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	if node == nil && db.tail() != nil {
		if node, err = db.tail().read(key); err != nil {
			return nil, err
		}
	}

	if node == nil {
		return nil, ErrNotFound
	}

	if node.tipe != TypeDict {
		return nil, ErrInvalidType
	}

	v, ok := node.dict[idx]
	if !ok {
		return nil, ErrInvalidIndex
	}

	return []byte(v), err
}

// ReadDict returns whole dict data
func (db *DB) ReadDict(key string) (map[string]string, error) {

	node, err := db.head().read(key)
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	if node == nil && db.tail() != nil {
		if node, err = db.tail().read(key); err != nil {
			return nil, err
		}
	}

	if node == nil {
		return nil, ErrNotFound
	}

	if node.tipe != TypeDict {
		return nil, ErrInvalidType
	}

	return node.dict, err
}

// Exists checking key existing
func (db *DB) Exists(key string) (bool, error) {

	val, err := db.head().read(key)

	return val != nil, err
}
