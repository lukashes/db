package main

import (
	"log"

	"github.com/lukashes/db/handler"
	"github.com/valyala/fasthttp"
)

const (
	addr = ":8080"
)

func main() {
	log.Printf("Started on %s", addr)
	fasthttp.ListenAndServe(addr, handler.Router)
}
