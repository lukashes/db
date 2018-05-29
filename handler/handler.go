package handler

import (
	"bytes"
	"strconv"

	"encoding/json"
	"github.com/labstack/gommon/log"
	"github.com/lukashes/db/db"
	"github.com/valyala/fasthttp"
)

var DB *db.DB

func init() {
	DB = db.New()
}

func Router(ctx *fasthttp.RequestCtx) {

	path := bytes.SplitN(bytes.Trim(ctx.Path(), "/"), []byte("/"), 4)

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
		return
	case "hget":
		d, err := DB.Read(string(path[2]))
		if err != nil {
			switch err {
			case db.ErrNotFound:
				ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
			default:
				log.Errorf("hget: %s", err)
				ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
			}
			return
		}
		ctx.Write(d)
	case "hset":
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
	case "rm":
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
	case "lset":
		var d []string
		if err := json.Unmarshal(ctx.PostBody(), &d); err != nil {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusBadRequest)
			return
		}
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
		if err := DB.WriteList(string(path[2]), d, ttl); err != nil {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}
	case "ladd":
		ctx.Response.Header.SetStatusCode(fasthttp.StatusNotImplemented)
	case "lget":
		// Get element by index
		if len(path) == 4 {
			i, err := strconv.Atoi(string(path[3]))
			if err != nil {
				ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
				return
			}
			d, err := DB.ReadListIndex(string(path[2]), i)
			if err != nil {
				switch err {
				case db.ErrNotFound, db.ErrInvalidIndex:
					ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
				case db.ErrInvalidType:
					ctx.Response.Header.SetStatusCode(fasthttp.StatusConflict)
				default:
					log.Errorf("lget: %s", err)
					ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
				}
				return
			}
			ctx.Write(d)
		} else { // Get whole list
			l, err := DB.ReadList(string(path[2]))
			if err != nil {
				switch err {
				case db.ErrNotFound, db.ErrInvalidIndex:
					ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
				case db.ErrInvalidType:
					ctx.Response.Header.SetStatusCode(fasthttp.StatusConflict)
				default:
					log.Errorf("lget: %s", err)
					ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
				}
				return
			}
			d, err := json.Marshal(l)
			if err != nil {
				log.Errorf("lget: %s", err)
				ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
				return
			}
			ctx.Write(d)
		}
	case "dset":
		d := map[string]string{}
		if err := json.Unmarshal(ctx.PostBody(), &d); err != nil {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusBadRequest)
			return
		}
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
		if err := DB.WriteDict(string(path[2]), d, ttl); err != nil {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}
	case "dadd":
		ctx.Response.Header.SetStatusCode(fasthttp.StatusNotImplemented)
		return
	case "dget":
		// Get element by index
		if len(path) == 4 {
			d, err := DB.ReadDictIndex(string(path[2]), string(path[3]))
			if err != nil {
				switch err {
				case db.ErrNotFound, db.ErrInvalidIndex:
					ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
				case db.ErrInvalidType:
					ctx.Response.Header.SetStatusCode(fasthttp.StatusConflict)
				default:
					log.Errorf("dget: %s", err)
					ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
				}
				return
			}
			ctx.Write(d)
		} else { // Get whole list
			l, err := DB.ReadDict(string(path[2]))
			if err != nil {
				switch err {
				case db.ErrNotFound, db.ErrInvalidIndex:
					ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
				case db.ErrInvalidType:
					ctx.Response.Header.SetStatusCode(fasthttp.StatusConflict)
				default:
					log.Errorf("dget: %s", err)
					ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
				}
				return
			}
			d, err := json.Marshal(l)
			if err != nil {
				log.Errorf("dget: %s", err)
				ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
				return
			}
			ctx.Write(d)
		}
	}

	ctx.Response.Header.SetStatusCode(fasthttp.StatusOK)
}
