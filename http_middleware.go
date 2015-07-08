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
	//固定参数名称
	_PARAM_FIELDS  = "fields"
	_PARAM_PAGE    = "page"
	_PARAM_PERPAGE = "per_page"
	_PARAM_DATE    = "date"
	_PARAM_START   = "start"
	_PARAM_END     = "end"
	_PARAM_ORDERBY = "orderby"

	//特殊前缀
	_PPREFIX_NOT  = '!'
	_PPREFIX_LIKE = '~'
	_PPREFIX_GT   = '>'
	_PPREFIX_LT   = '<'

	OriginalRemoteAddrKey = "originalRemoteAddr"

	// 查询类型
	CTYPE_IS = iota
	CTYPE_NOT
	CTYPE_LIKE
	CTYPE_GT
	CTYPE_LT
	CTYPE_JOIN
	CTYPE_RANGE
	CTYPE_ORDER
	CTYPE_PAGE
	CTYPE_RAW
)

var (
	xForwardedFor      = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP            = http.CanonicalHeaderKey("X-Real-IP")
	contentType        = http.CanonicalHeaderKey("Content-Type")
	contentDisposition = http.CanonicalHeaderKey("Content-Disposition")
	contentMD5         = http.CanonicalHeaderKey("Content-MD5")
	rcHolder           func(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext
)

/* {{{ func getCTypeByPrefix(p string) int
 *
 */
func getCTypeByPrefix(p byte) int {
	switch p {
	case _PPREFIX_NOT:
		return CTYPE_NOT
	case _PPREFIX_LIKE:
		return CTYPE_LIKE
	case _PPREFIX_GT:
		return CTYPE_GT
	case _PPREFIX_LT:
		return CTYPE_LT
	default:
		return CTYPE_IS
	}
}

/* }}} */

/* {{{ func EnvInit(c *web.C, h http.Handler) http.Handler
 * 初始化环境
 */
func EnvInit(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ac := new(Access) //access日志信息
		ac.Time = time.Now()
		ac.Http = new(HTTPLog)
		ac.Http.Method = r.Method
		ac.Http.URI = r.RequestURI
		ac.Http.Proto = r.Proto
		ac.Http.Host = r.Host
		ac.Http.InHeader = &r.Header
		// env
		if c.Env == nil {
			//c.Env = make(map[string]interface{})
			c.Env = make(map[interface{}]interface{})
		}

		// make rand string(for debug, session...)
		buf := make([]byte, 16)
		randbo.New().Read(buf) //号称最快的随机字符串
		ac.Session = fmt.Sprintf("%x", buf)

		c.Env[RequestIDKey] = ac.Session

		c.Env[LogPrefixKey] = "[" + ac.Session[:10] + "]" //只显示前十位

		Trace("[%s] [%s %s] started", ac.Session[:10], r.Method, r.RequestURI)

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
		ac.Http.IP = r.RemoteAddr

		//init RESTContext
		var rcErr error
		var rc *RESTContext
		rc, rcHolder, rcErr = RCHolder(*c, w, r)
		rc.Access = ac
		rc.Access.Http.ReqLength = len(rc.RequestBody)
		if rcErr != nil {
			rc.RESTBadRequest(rcErr)
			return
		}

		h.ServeHTTP(lw, r)

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
		defer func() {
			if err := recover(); err != nil {
				rc.Critical("[%s %s] %v", r.Method, r.URL.Path, err)
				//debug.PrintStack()
				rc.Critical("%s", debug.Stack())
				//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				rc.HTTPError(http.StatusInternalServerError)
			}

			// save access log here
			ac := rc.Access
			ac.Duration = time.Now().Sub(ac.Time).String()
			ac.Http.Status = rc.Status
			ac.Http.OutHeader = w.Header()
			ac.Http.RepLength = rc.ContentLength
			if sb := rc.GetEnv(SaveBodyKey); sb != nil && sb.(bool) == true {
				//可以由应用程序决定是否记录body
				ac.Http.ReqBody = string(rc.RequestBody)
			}
			//ac.App = string(rc.RequestBody)

			rc.Debug("[%s %s] end:%d in %s", ac.Http.Method, ac.Http.URI, ac.Http.Status, ac.Duration)
			// save access
			rc.SaveAccess()
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
		var ct int
		var p, pp string
		rc.setTimeRangeFromStartEnd()
		for k, v := range r.Form {
			switch k { //处理参数
			case _PARAM_DATE:
				rc.setTimeRangeFromDate(v)
			case _PARAM_ORDERBY:
				rc.setOrderBy(v)
			case _PARAM_FIELDS:
				//过滤字段
				if len(v) > 1 { //传了多个
					rc.SetEnv(FieldsKey, v)
				} else {
					if strings.Contains(v[0], ",") {
						rc.SetEnv(FieldsKey, strings.Split(v[0], ","))
					} else {
						rc.SetEnv(FieldsKey, v)
					}
				}
			case _PARAM_PERPAGE:
				if len(v) > 0 {
					pp = v[0]
				}
			case _PARAM_PAGE: //分页信息
				if len(v) > 0 {
					p = v[0]
				}
			default:
				//除了以上的特别字段,其他都是条件查询
				var cv interface{}
				//var con *Condition
				//var err error

				if len(v) > 1 {
					cv = v
				} else {
					//cv = v[0]
					//处理逗号情况
					if strings.Contains(v[0], ",") {
						cv = strings.Split(v[0], ",")
					} else {
						cv = v[0]
					}
				}

				//根据参数名第一个字符来判断条件类型
				prefix := k[0] //param prefix
				if ct = getCTypeByPrefix(prefix); ct != CTYPE_IS {
					k = k[1:]
					//Debug("[key: %s][ctype: %d]", k, ct)
				}

				//如果参数中包含".",代表有关联查询
				if strings.Contains(k, ".") {
					js := strings.SplitN(k, ".", 2)
					if js[0] != "" && js[1] != "" {
						k = js[0]
						cv = NewCondition(ct, js[1], cv)
						//查询类型变为join
						rc.Trace("join: %s, %s; con: %v", k, cv.(*Condition).Field, cv)
						ct = CTYPE_JOIN
					}
				}

				rc.setCondition(NewCondition(ct, k, cv))
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
