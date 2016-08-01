package client

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"
)

const (
	Version = "v1"
)

type Value struct {
	v []byte
	b bytes.Buffer
}

func (v *Value) Read(p []byte) (n int, err error) {
	return v.b.Read(p)
}

func (v *Value) Decode(r interface{}) error {
	d := gob.NewDecoder(v)
	return d.Decode(r)
}

func valueFromMap(m map[string]interface{}) (*Value, error) {
	b := new(bytes.Buffer)

	e := gob.NewEncoder(b)

	err := e.Encode(m)
	if err != nil {
		return nil, err
	}

	return &Value{v: b.Bytes(), b: *b}, nil
}

func valueFromList(l []interface{}) (*Value, error) {
	b := new(bytes.Buffer)

	e := gob.NewEncoder(b)

	err := e.Encode(l)
	if err != nil {
		return nil, err
	}

	return &Value{v: b.Bytes(), b: *b}, nil
}

func valueFromInt(i int64) (*Value, error) {
	b := new(bytes.Buffer)

	e := gob.NewEncoder(b)

	err := e.Encode(i)
	if err != nil {
		return nil, err
	}

	return &Value{v: b.Bytes(), b: *b}, nil
}

func valueFromString(s string) (*Value, error) {
	b := new(bytes.Buffer)

	e := gob.NewEncoder(b)

	err := e.Encode(s)
	if err != nil {
		return nil, err
	}

	return &Value{v: b.Bytes(), b: *b}, nil
}

type DB struct {
	mu     sync.Mutex
	client *http.Client
	prefix string
}

func New(addr string) (*DB, error) {
	return &DB{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 10,
			},
		},
		prefix: addr + "/" + Version,
	}, nil
}

func (db *DB) Get(key string) (*Value, error) {
	res, err := db.client.Get(db.prefix + "/get/" + key)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return &Value{v: body, b: *bytes.NewBuffer(body)}, nil
}

func (db *DB) Add(key string, val interface{}) error {
	var v *Value

	switch t := val.(type) {
	default:
		return fmt.Errorf("Unexpected type")
	case map[string]interface{}:
		v, _ = valueFromMap(t)
	case []interface{}:
		v, _ = valueFromList(t)
	case int, int8, int16, int32, int64:
		v, _ = valueFromInt(reflect.ValueOf(t).Int())
	case string:
		v, _ = valueFromString(t)
	}

	res, err := db.client.Post(db.prefix+"/set/"+key, "application/x-www-form-urlencoded", v)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func (db *DB) Remove(key string) error {
	res, err := db.client.Get(db.prefix + "/remove/" + key)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
