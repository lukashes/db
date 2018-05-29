package db

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

func initKeys(n int) (keys []string) {
	for k := range rand.Perm(n) {
		keys = append(keys, strconv.Itoa(k))
	}

	return keys
}

func BenchmarkMapWrite(b *testing.B) {
	db := make(map[string][]byte)
	keys := initKeys(b.N)

	b.ResetTimer()
	for _, k := range keys {
		db[k] = []byte(k)
	}
}

func BenchmarkMapRead(b *testing.B) {
	db := make(map[string][]byte)
	keys := initKeys(b.N)
	check := initKeys(b.N)
	for _, k := range keys {
		db[k] = []byte(k)
	}

	b.ResetTimer()
	for _, k := range check {
		_ = db[k]
	}
}

func BenchmarkSyncMapWrite(b *testing.B) {
	db := sync.Map{}
	keys := initKeys(b.N)

	b.ResetTimer()
	for _, k := range keys {
		db.Store(k, []byte(k))
	}
}

func BenchmarkSyncMapRead(b *testing.B) {
	db := sync.Map{}
	keys := initKeys(b.N)
	check := initKeys(b.N)
	for _, k := range keys {
		db.Store(k, []byte(k))
	}

	b.ResetTimer()
	for _, k := range check {
		db.Load(k)
	}
}

func BenchmarkHashTableWrite(b *testing.B) {
	db := New()
	keys := initKeys(b.N)

	b.ResetTimer()
	for _, k := range keys {
		db.Write(k, []byte(k), nil)
	}
}

func BenchmarkHashTableRead(b *testing.B) {
	db := New()
	keys := initKeys(b.N)
	check := initKeys(b.N)
	for _, k := range keys {
		db.Write(k, []byte(k), nil)
	}

	b.ResetTimer()
	for _, k := range check {
		db.Read(k)
	}
}

func BenchmarkHashTableWriteList(b *testing.B) {
	db := New()
	data := map[int][]string{
		0: {"1", "2", "3"},
		1: {"1"},
		2: {"qwe", "wer", "ert"},
	}
	keys := initKeys(b.N)

	b.ResetTimer()
	for i, k := range keys {
		db.WriteList(k, data[i%3], nil)
	}
}

func BenchmarkHashTableListItemRead(b *testing.B) {
	db := New()
	data := map[int][]string{
		0: {"1", "2", "3"},
		1: {"1"},
		2: {"qwe", "wer", "ert"},
	}
	keys := initKeys(b.N)
	check := initKeys(b.N)

	for i, k := range keys {
		db.WriteList(k, data[i%3], nil)
	}

	b.ResetTimer()
	for _, k := range check {
		db.ReadListIndex(k, 3)
	}
}

func TestConsistency(t *testing.T) {
	db := New()

	key := "test"
	val := []byte("val")

	if err := db.Write(key, val, nil); err != nil {
		t.Error(err)
	}

	new, err := db.Read(key)
	if err != nil {
		t.Error(err)
	}

	if string(val) != string(new) {
		t.Fail()
	}
}

func TestList(t *testing.T) {
	db := New()

	key := "test"
	expected := []string{"donald", "duck"}

	if err := db.WriteList(key, expected, nil); err != nil {
		t.Error(err)
	}

	for i := range expected {
		v, err := db.ReadListIndex(key, i)
		if err != nil {
			t.Error(err)
		}

		if string(v) != expected[i] {
			t.Error("invalid value")
		}
	}

	/*for k, b := range db.head().buckets {
		fmt.Printf("bucket %d: %#v\n", k, *b)
	}*/
}

func TestGrowing(t *testing.T) {
	db := New()
	count := 100000
	keys := initKeys(count)
	check := initKeys(count)

	for _, k := range keys {
		db.Write(k, []byte(k), nil)
	}

	/*for k, b := range db.head().buckets {
		fmt.Printf("before growing bucket %d nodes:\n", k)
		for n := b.nodes; n != nil; n = n.next {
			fmt.Printf("	%#v\n", *n)
		}
	}*/

	//time.Sleep(time.Second)

	for _, k := range check {
		r, err := db.Read(k)
		if err != nil {
			t.Errorf("unexpected error for key %s: %s", k, err)
		} else if k != string(r) {
			t.Errorf("invalid value %s, expected %s", string(r), k)
		}
	}

	/*time.Sleep(time.Second)

	for k, b := range db.head().buckets {
		fmt.Printf("bucket %d nodes:\n", k)
		for n := b.nodes; n != nil; n = n.next {
			fmt.Printf("	%#v\n", *n)
		}
	}*/
}
