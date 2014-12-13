package ogo

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Odinman/ogo/utils"
	"github.com/dustin/randbo"
	"github.com/zenazn/goji/web"
)

// Key to use when setting the request ID.
const (
	RequestIDKey          = "_reqID"
	LogPrefixKey          = "_prefix"
	OriginalRemoteAddrKey = "originalRemoteAddr"
)

var (
	xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP       = http.CanonicalHeaderKey("X-Real-IP")
	rcHolder      func(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext
)

//初始化环境
func EnvInit(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		// env
		if c.Env == nil {
			c.Env = make(map[string]interface{})
		}

		// make rand string(for debug, session...)
		buf := make([]byte, 16)
		randbo.New().Read(buf) //号称最快的随即字符串
		reqID := fmt.Sprintf("%x", buf)

		c.Env[RequestIDKey] = reqID

		c.Env[LogPrefixKey] = "[" + reqID[:10] + "]" //只显示前十位

		Debug("[%s] [%s %s] started", reqID[:10], r.Method, r.URL.Path)

		lw := utils.WrapWriter(w)

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

		//init RESTContext
		rcHolder = RCHolder(*c, w, r)

		h.ServeHTTP(lw, r)

		if lw.Status() == 0 {
			lw.WriteHeader(http.StatusOK)
		}
		t2 := time.Now()

		Debug("[%s] [%s %s] end:%d in %s", reqID[:10], r.Method, r.URL.Path, lw.Status(), t2.Sub(t1))
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

		rc := rcHolder(*c, w, r)
		//Debug("defer len: %d", len(rc.RequestBody))
		defer func() {
			if err := recover(); err != nil {
				rc.Criticalf("[%s %s] %v", r.Method, r.URL.Path, err)
				debug.PrintStack()
				//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				rc.HTTPError(http.StatusInternalServerError)
			}

			// save access log here
		}()

		h.ServeHTTP(w, r)
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
