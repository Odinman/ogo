package ogo

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/util"
)

// Key to use when setting the request ID.
const RequestIDKey = "reqID"

// Defer is a middleware that recovers from panics, logs the panic (and a
// backtrace), and returns a HTTP 500 (Internal Server Error) status if
// possible.
// save access log
// Recoverer prints a request ID if one is provided.
func Defer(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		reqID := GetReqID(*c)

		defer func() {
			if err := recover(); err != nil {
				//printPanic(reqID, err)
				Debugger.Critical("[%s][url: %s] %v", reqID, r.URL.Path, err)
				debug.PrintStack()
				http.Error(w, http.StatusText(500), 500)
			}

			// save access log here
		}()

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

//初始化环境
func EnvInit(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		reqID := GetReqID(*c)

		Debugger.Debug("[%s][url: %s] started", reqID, r.URL.Path)

		lw := util.WrapWriter(w)

		t1 := time.Now()
		h.ServeHTTP(lw, r)

		if lw.Status() == 0 {
			lw.WriteHeader(http.StatusOK)
		}
		t2 := time.Now()

		Debugger.Debug("[%s][url: %s] end:%d in %s", reqID, r.URL.Path, lw.Status(), t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

// GetReqID returns a request ID from the given context if one is present.
// Returns the empty string if a request ID cannot be found.
func GetReqID(c web.C) string {
	if c.Env == nil {
		return ""
	}
	v, ok := c.Env[RequestIDKey]
	if !ok {
		return ""
	}
	if reqID, ok := v.(string); ok {
		return reqID
	}
	return ""
}
