// Ogo

package ogo

import (
	//"fmt"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"net/http"
	"strings"
)

const (
	_OPS_SKIPAUTH = iota
	_OPS_SKIPLOGIN
)

type Handler func(c *RESTContext)

type Route struct {
	Pattern   string
	Method    string
	Handler   Handler
	SkipAuth  bool //是否忽略鉴权(默认需要)
	SkipLogin bool //是否忽略登录
}

type Controller struct {
	Endpoint string
	Routes   map[string]*Route
	ReqCount int //访问计数
	Mux      *Mux
}

type ControllerInterface interface {
	Init(endpoint string, c ControllerInterface)
	SetMux(mux *Mux)
	Get(c *RESTContext)
	Post(c *RESTContext)
	Put(c *RESTContext)
	Delete(c *RESTContext)
	Patch(c *RESTContext)
	Head(c *RESTContext)
	Options(c *RESTContext)
	Trace(c *RESTContext)
	NotFound(c *RESTContext)
	AddRoute(m string, p string, h Handler, options ...bool)
}

func NewRoute(p string, m string, h Handler, options ...bool) *Route {
	r := &Route{
		Pattern: p,
		Method:  m,
		Handler: h,
	}

	if len(options) > 0 {
		for offset, option := range options {
			switch offset {
			case _OPS_SKIPAUTH:
				r.SkipAuth = option
			case _OPS_SKIPLOGIN:
				r.SkipLogin = option
			default:
				// nothing to do
			}
		}
	}

	return r
}

// 封装
//func handlerWrap(f Handler) web.HandlerFunc { //这里封装了webC到本地的结构中
func handlerWrap(rt *Route) web.HandlerFunc { //这里封装了webC到本地的结构中
	fn := func(c web.C, w http.ResponseWriter, r *http.Request) {
		// build newest RESTContext
		rc := rcHolder(c, w, r)

		//route
		rc.Route = rt

		//Debug("method: %s, bodylen: %d", r.Method, len(rc.RequestBody))

		// pre hooks, 任何一个出错,都要结束
		if hl := len(DMux.Hooks.preHooks); hl > 0 {
			for i := 0; i < hl; i++ {
				if err := DMux.Hooks.preHooks[i](rc); err != nil {
					rc.RESTError(err)
					return
				}
			}
		}

		// 执行业务handler
		rt.Handler(rc)

		// post hooks
		if hl := len(DMux.Hooks.postHooks); hl > 0 {
			for i := 0; i < hl; i++ {
				DMux.Hooks.postHooks[i](rc)
			}
		}
	}
	return fn
}

/* {{{ func (ctr *Controller) SetMux(mux *Mux)
 *
 */
func (ctr *Controller) SetMux(mux *Mux) {
	ctr.Mux = mux
}

/* }}} */

func (ctr *Controller) Init(endpoint string, c ControllerInterface) {
	ctr.Endpoint = endpoint
	//ctr.Routes = make(map[string]*Route)
	//默认路由
	ctr.DefaultRoutes(c)
	if len(ctr.Routes) > 0 {
		for key, rt := range ctr.Routes {
			//Debug("pattern: %s", rt.Pattern)
			// regist routes to Mux
			ctr.Mux.Routes[key] = rt
			switch strings.ToLower(rt.Method) {
			case "get":
				ctr.RouteGet(rt)
			case "post":
				ctr.RoutePost(rt)
			case "put":
				ctr.RoutePut(rt)
			case "delete":
				ctr.RouteDelete(rt)
			case "patch":
				ctr.RoutePatch(rt)
			case "head":
				ctr.RouteHead(rt)
			default:
				// unknow method
			}
		}
	}
	// not found
	notFoundRoute := &Route{
		Handler: c.NotFound,
	}
	ctr.RouteNotFound(notFoundRoute)
}

func (ctr *Controller) Get(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (ctr *Controller) Post(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (ctr *Controller) Put(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (ctr *Controller) Delete(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (ctr *Controller) Patch(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (ctr *Controller) Head(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (ctr *Controller) Options(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (ctr *Controller) Trace(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (ctr *Controller) NotFound(c *RESTContext) {
	c.HTTPError(http.StatusNotFound)
}

func (ctr *Controller) AddRoute(m string, p string, h Handler, options ...bool) {
	key := strings.ToUpper(m) + " " + p
	if ctr.Routes == nil {
		ctr.Routes = make(map[string]*Route)
	}
	if _, ok := ctr.Routes[key]; ok {
		//手动加路由, 以最后加的为准,overwrite
	}
	ctr.Routes[key] = NewRoute(p, m, h, options...)
}

// controller default route
// 默认路由, 如果已经定义了则忽略，没有定义则加上
//func (ctr *Controller) DefaultRoutes() {
func (ctr *Controller) DefaultRoutes(c ControllerInterface) {
	var pattern, method, key string
	// GET /{endpoint}
	pattern = "/" + ctr.Endpoint
	method = "GET"
	key = method + " " + pattern
	if ctr.Routes == nil {
		ctr.Routes = make(map[string]*Route)
	}
	if _, ok := ctr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
	} else {
		rt := NewRoute(pattern, method, c.Get)
		ctr.Routes[key] = rt
	}

	// GET /{endpoint}/{id}
	pattern = "/" + ctr.Endpoint + "/:_id_"
	method = "GET"
	key = method + " " + pattern
	if _, ok := ctr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
	} else {
		rt := NewRoute(pattern, method, c.Get)
		ctr.Routes[key] = rt
	}

	// POST /{endpoint}
	pattern = "/" + ctr.Endpoint
	method = "POST"
	key = method + " " + pattern
	if _, ok := ctr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
	} else {
		rt := NewRoute(pattern, method, c.Post)
		ctr.Routes[key] = rt
	}

	// DELETE /{endpoint}/{id}
	pattern = "/" + ctr.Endpoint + "/:_id_"
	method = "DELETE"
	key = method + " " + pattern
	if _, ok := ctr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
	} else {
		rt := NewRoute(pattern, method, c.Delete)
		ctr.Routes[key] = rt
	}

	// PATCH /{endpoint}/{id}
	pattern = "/" + ctr.Endpoint + "/:_id_"
	method = "PATCH"
	key = method + " " + pattern
	if _, ok := ctr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
	} else {
		rt := NewRoute(pattern, method, c.Patch)
		ctr.Routes[key] = rt
	}

	// PUT /{endpoint}/{id}
	pattern = "/" + ctr.Endpoint + "/:_id_"
	method = "PUT"
	key = method + " " + pattern
	if _, ok := ctr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
	} else {
		rt := NewRoute(pattern, method, c.Put)
		ctr.Routes[key] = rt
	}
}

func (ctr *Controller) RouteGet(rt *Route) {
	goji.Get(rt.Pattern, handlerWrap(rt))
}

func (ctr *Controller) RoutePost(rt *Route) {
	goji.Post(rt.Pattern, handlerWrap(rt))
}

func (ctr *Controller) RoutePut(rt *Route) {
	goji.Put(rt.Pattern, handlerWrap(rt))
}

func (ctr *Controller) RouteDelete(rt *Route) {
	goji.Delete(rt.Pattern, handlerWrap(rt))
}

func (ctr *Controller) RoutePatch(rt *Route) {
	goji.Patch(rt.Pattern, handlerWrap(rt))
}

func (ctr *Controller) RouteHead(rt *Route) {
	goji.Head(rt.Pattern, handlerWrap(rt))
}

func (ctr *Controller) RouteNotFound(rt *Route) {
	goji.NotFound(handlerWrap(rt))
}
