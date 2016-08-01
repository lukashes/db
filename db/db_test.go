package db_test

import (
	"strconv"
	"testing"

	"github.com/lukashes/db/db"
)

func BenchmarkHashTableWrite(b *testing.B) {
	db := db.New()
	for i := 0; i < b.N; i++ {
		kv := strconv.Itoa(i)
		db.Write(kv, []byte(kv), nil)
	}
}

func BenchmarkHashTableRead(b *testing.B) {
	db := db.New()
	for i := 0; i < b.N; i++ {
		kv := strconv.Itoa(i)
		db.Write(kv, []byte(kv), nil)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kv := strconv.Itoa(i)
		db.Read(kv)
	}
}

func TestConsistency(t *testing.T) {
	db := db.New()

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
