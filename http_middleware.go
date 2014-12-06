package ogo

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Odinman/ogo/utils"
	"github.com/zenazn/goji/web"
)

// Key to use when setting the request ID.
const RequestIDKey = "reqID"
const OriginalRemoteAddrKey = "originalRemoteAddr"

var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

//初始化环境
func EnvInit(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		reqID := GetReqID(*c)

		t1 := time.Now()
		Debugger.Debug("[%s][url: %s] started", reqID, r.URL.Path)

		lw := utils.WrapWriter(w)

		//new rest context
		RESTC = newContext(*c, lw, r)

		//request body
		//if CopyRequestBody && r.Method != "GET" && r.Method != "HEAD" && r.Method!= "DELETE"{
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "DELETE" {
			RESTC.RequestBody, _ = ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			bf := bytes.NewBuffer(RESTC.RequestBody)
			r.Body = ioutil.NopCloser(bf)
		}

		// 解析参数
		r.ParseForm()

		// env
		if c.Env == nil {
			c.Env = make(map[string]interface{})
		}
		pathPieces := strings.Split(r.URL.Path, "/")
		for off, piece := range pathPieces {
			if piece != "" {
				if off == 1 {
					c.Env["_endpoint"] = piece
				}
				if off == 2 && piece[0] != '@' { //@开头是selector
					c.Env["_rowkey"] = piece
				}
				if off > 1 && piece[0] == '@' {
					c.Env["_selector"] = piece
				}
			}
		}
		// real ip(处理在代理服务器之后的情况)
		if rip := realIP(r); rip != "" {
			c.Env[OriginalRemoteAddrKey] = r.RemoteAddr
			r.RemoteAddr = rip
		}

		h.ServeHTTP(lw, r)

		if lw.Status() == 0 {
			lw.WriteHeader(http.StatusOK)
		}
		t2 := time.Now()

		Debugger.Debug("[%s][url: %s] end:%d in %s", reqID, r.URL.Path, lw.Status(), t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

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
				//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				RESTC.HTTPError(http.StatusInternalServerError)
			}

			// save access log here
		}()

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

//Authentication
func Authentication(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

/* {{{ func RunHooks(c *web.C, h http.Handler) http.Handler
 * 处理钩子函数
 */
func RunHooks(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		if hl := len(Hooks.preHooks); hl > 0 {
			for i := 0; i < hl; i++ {
				Hooks.preHooks[i](getContext(*c, w, r))
			}
		}

		h.ServeHTTP(w, r)

		if hl := len(Hooks.postHooks); hl > 0 {
			for i := 0; i < hl; i++ {
				Hooks.postHooks[i](getContext(*c, w, r))
			}
		}
	}

	return http.HandlerFunc(fn)
}

/* }}} */

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

func realIP(r *http.Request) string {
	var ip string

	if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	}

	return ip
}
