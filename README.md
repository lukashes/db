# DB

Simple in-memory DB with dynamic hash table storage

## Running

Get it
```
go get -u github.com/lukashes/db
```

After getting run HTTP-server
```
db$ go run main.go
2016/08/01 19:00:33 Started on :8080
```

Or start container
```
docker build -t lukashes/db:latest .
docker run -p 8080:8080 --rm -it lukashes/db:latest
```

Now you can create your HTTP requests

## HTTP API

Params:

* key - should be alphanumeric (not validated yet)
* value - should be in the body
* ttl - time to live in seconds

### POST /v1/hset/key?ttl=seconds

Set value by the key. If key exists it will be rewritten.

### GET /v1/hget/key

Read value by the key

### POST /v1/lset/key?ttl=seconds

Set list by the key. List should be provided in body as json.

### GET /v1/lget/key/index

Read value placed at list index. If index not provided returns whole list as json.

### POST /v1/dset/key?ttl=seconds

Set dict by the key. Dict should be provided in body as json.

### GET /v1/dget/key/index

Read value placed at dict index. If index not provided returns whole dict as json.

### GET /v1/rm/key

Remove value by the key

### GET /v1/keys

Get all available keys

Response will receive all keys separated by comma

## Client

Now it works only by HTTP, sorry...

Example of client here client/example/main.go

## Benchmarks

```
db$ go test -bench . -benchmem ./db/...
  goos: linux
  goarch: amd64
  pkg: github.com/lukashes/db/db
  BenchmarkMapWrite-4                      2000000               686 ns/op             208 B/op          1 allocs/op
  BenchmarkMapRead-4                      10000000               168 ns/op               0 B/op          0 allocs/op
  BenchmarkSyncMapWrite-4                  1000000              1528 ns/op             219 B/op          6 allocs/op
  BenchmarkSyncMapRead-4                  10000000               330 ns/op               0 B/op          0 allocs/op
  BenchmarkHashTableWrite-4                2000000              1110 ns/op             193 B/op          4 allocs/op
  BenchmarkHashTableRead-4                 3000000               632 ns/op               0 B/op          0 allocs/op
  BenchmarkHashTableWriteList-4            2000000               901 ns/op             185 B/op          3 allocs/op
  BenchmarkHashTableListItemRead-4         3000000               578 ns/op               0 B/op          0 allocs/op
  PASS
```

## TODO

* TCP connection
* Releasing memory
* Improve client
* Write more tests