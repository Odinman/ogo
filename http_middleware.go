package ogo

import (
	"fmt"
	"mime"
	"net/http"
	//"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Odinman/ogo/utils"
	"github.com/dustin/randbo"
	"github.com/zenazn/goji/web"
)

// Key to use when setting the request ID.
const (
	_PARAM_FIELDS         = "fields"
	_PARAM_PAGE           = "page"
	_PARAM_PERPAGE        = "per_page"
	_PPREFIX_NOT          = '!'
	_PPREFIX_LIKE         = '~'
	_CTYPE_IS             = 0
	_CTYPE_NOT            = 1
	_CTYPE_LIKE           = 2
	OriginalRemoteAddrKey = "originalRemoteAddr"
)

var (
	xForwardedFor      = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP            = http.CanonicalHeaderKey("X-Real-IP")
	contentType        = http.CanonicalHeaderKey("Content-Type")
	contentDisposition = http.CanonicalHeaderKey("Content-Disposition")
	contentMD5         = http.CanonicalHeaderKey("Content-MD5")
	rcHolder           func(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext
)

/* {{{ func EnvInit(c *web.C, h http.Handler) http.Handler
 * 初始化环境
 */
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
					c.Env[EndpointKey] = piece
				}
				if off == 2 && piece[0] != '@' { //@开头是selector
					c.Env[RowkeyKey] = piece
				}
				if off > 1 && piece[0] == '@' {
					c.Env[SelectorKey] = piece
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

/* }}} */

/* {{{ func Defer(c *web.C, h http.Handler) http.Handler
 * recovers from panics
 */
func Defer(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		rc := rcHolder(*c, w, r)
		//Debug("defer len: %d", len(rc.RequestBody))
		defer func() {
			if err := recover(); err != nil {
				rc.Critical("[%s %s] %v", r.Method, r.URL.Path, err)
				//debug.PrintStack()
				rc.Critical("%s", debug.Stack())
				//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				rc.HTTPError(http.StatusInternalServerError)
			}

			// save access log here
		}()

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

/* }}} */

/* {{{ func Mime(c *web.C, h http.Handler) http.Handler
 * mimetype相关处理
 */
func Mime(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		rc := rcHolder(*c, w, r)

		if cs := r.Header.Get(contentMD5); cs != "" {
			rc.SetEnv(ContentMD5Key, cs)
		}

		// 看content-type
		if ct := r.Header.Get(contentType); ct != "" {
			rc.SetEnv(MimeTypeKey, ct)
		}
		if cd := r.Header.Get(contentDisposition); cd != "" {
			//以传入的Disposition为主
			if t, m, e := mime.ParseMediaType(cd); e == nil {
				rc.Info("disposition: %s, mediatype: %s", cd, t)
				rc.SetEnv(DispositionMTKey, t)
				//if fname, ok := m["filename"]; ok {
				//	if mt := mime.TypeByExtension(filepath.Ext(fname)); mt != "" {
				//		rc.SetEnv(MimeTypeKey, mt)
				//	}
				//}
				for k, v := range m {
					dk := DispositionPrefix + k + "_"
					rc.Debug("disposition key: %s, value: %v", dk, v)
					rc.SetEnv(dk, v)
				}
			}

		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

/* }}} */

/* {{{ func ParseParams(c *web.C, h http.Handler) http.Handler {
 *
 */
func ParseParams(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		rc := rcHolder(*c, w, r)
		// 解析参数
		r.ParseForm()
		// 根据ogo规则解析参数
		var cType int
		var p, pp string
		for k, v := range r.Form {
			rc.Trace("key: %s, value: %s", k, v)
			//根据参数名第一个字符来判断条件类型
			prefix := k[0] //param prefix
			switch prefix {
			case _PPREFIX_NOT:
				rc.Debug("having prefix not: %s", k)
				k = k[1:]
				cType = _CTYPE_NOT
				rc.Debug("key change to: %s, condition type: %d", k, cType)
			case _PPREFIX_LIKE:
				k = k[1:]
				cType = _CTYPE_LIKE
			default:
				cType = _CTYPE_IS
				// do nothing
			}

			switch k { //处理参数
			case _PARAM_FIELDS:
				//过滤字段
				rc.SetEnv(FieldsKey, v)
			case _PARAM_PERPAGE:
			case _PARAM_PAGE: //分页信息
				if len(v) > 0 {
					p = v[0]
				}
				if pps, ok := r.Form[_PARAM_PERPAGE]; ok {
					if len(pps) > 0 {
						pp = pps[0]
					}
				}
			default:
				//除了以上的特别字段,其他都是条件查询
				var con *Condition
				var err error
				if con, err = rc.GetCondition(k); err != nil {
					//没有这个condition,初始化
					con = new(Condition)
					rc.setCondition(k, con)
				}
				switch cType {
				case _CTYPE_IS:
					con.Is = v
				case _CTYPE_NOT:
					con.Not = v
				case _CTYPE_LIKE:
					con.Like = v
				default:
				}
				rc.Trace("con: %v", con)
			}
		}
		//记录分页信息
		rc.SetEnv(PaginationKey, NewPagination(p, pp))

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

/* }}} */

/* {{{ func realIP(r *http.Request) string
 * 获取真实IP
 */
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

/* }}} */
