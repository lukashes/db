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
* value - should be write/read from body
* ttl - time to live in seconds

### POST /v1/set/key?ttl=seconds

Set value to by the key. If key exists it will be rewritten.

Params:

### POST /v1/update/key

Update value only it exists.

### GET /v1/get/key

Read value by key

### GET /v1/remove/key

Remove value by key

### GET /v1/keys

Get all available keys

Response will receive all keys separated by comma

## Client

Now it works only by HTTP, sorry...

Example of client here client/example/main.go

## Benchmarks

```
db$ go test -bench . -benchmem ./db/...
PASS
BenchmarkHashTableWrite-4        1000000              1412 ns/op             380 B/op          3 allocs/op
BenchmarkHashTableRead-4         2000000               975 ns/op              71 B/op          2 allocs/op
```

## TODO

* TCP connection
* Releasing memory
* Improve client
* Write more tests