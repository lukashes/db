package handler

import (
	"bytes"
	"strconv"

	"github.com/lukashes/db/db"
	"github.com/valyala/fasthttp"
)

var DB *db.DB

func init() {
	DB = db.New()
}

func Router(ctx *fasthttp.RequestCtx) {

	path := bytes.SplitN(bytes.Trim(ctx.Path(), "/"), []byte("/"), 3)

	if len(path) < 2 {
		ctx.Response.Header.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	if !bytes.Equal(path[0], []byte("v1")) {
		ctx.Response.Header.SetStatusCode(fasthttp.StatusNotImplemented)
		return
	}

	switch string(path[1]) {
	default:
		ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
	case "get":
		d, err := DB.Read(string(path[2]))
		if err != nil {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}
		ctx.Write(d)
	case "set":
		d := ctx.PostBody()
		var ttl *int
		t := ctx.QueryArgs().Peek("ttl")
		if len(t) > 0 {
			tt, err := strconv.Atoi(string(t))
			if err != nil {
				ctx.Response.Header.SetStatusCode(fasthttp.StatusBadRequest)
				return
			}
			ttl = &tt
		}
		if err := DB.Write(string(path[2]), d, ttl); err != nil {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}
	case "update":
		exists, err := DB.Exists(string(path[2]))
		if err != nil {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}
		if !exists {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
			return
		}
		d := ctx.PostBody()
		DB.Write(string(path[2]), d, nil)
	case "remove":
		if err := DB.Delete(string(path[2])); err != nil {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
		}
	case "keys":
		keys := DB.Keys()
		for k, v := range keys {
			if k != 0 {
				ctx.Write([]byte(","))
			}
			ctx.Write([]byte(v))
		}
	}

	ctx.Response.Header.SetStatusCode(fasthttp.StatusOK)
}
