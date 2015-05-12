// Ogo

package ogo

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

const (
	// generic controller const
	GC_GET = 1 << iota
	GC_POST
	GC_DELETE
	GC_PATCH
	//GC_PUT
	GC_HEAD
	GC_ALL = GC_GET | GC_POST | GC_DELETE | GC_PATCH | GC_HEAD

	//KEY_SKIPAUTH  = "skipauth"
	//KEY_SKIPLOGIN = "skiplogin"
	//KEY_SKIPPERM  = "skipperm"
)

type Handler func(c *RESTContext)

type RouteOption map[string]interface{}

type Route struct {
	Pattern interface{}
	Method  string
	Handler Handler
	Options RouteOption
}

type Router struct {
	Endpoint string
	Routes   map[string]*Route
	SRoutes  []*Route //排序的Route
	ReqCount int      //访问计数
	Mux      *Mux
}

type RouterInterface interface {
	Init(c RouterInterface)
	New(mux *Mux, endpoint string)
	GetEndpoint() string
	Get(c *RESTContext)
	Post(c *RESTContext)
	Put(c *RESTContext)
	Delete(c *RESTContext)
	Patch(c *RESTContext)
	Head(c *RESTContext)
	Options(c *RESTContext)
	Trace(c *RESTContext)
	NotFound(c *RESTContext)
	AddRoute(m string, p interface{}, h Handler, options ...map[string]interface{})
}

/* {{{ func NewRoute(p interface{}, m string, h Handler, options ...map[string]interface{}) *Route
 *
 */
func NewRoute(p interface{}, m string, h Handler, options ...map[string]interface{}) *Route {
	r := &Route{
		Pattern: p,
		Method:  m,
		Handler: h,
		Options: make(map[string]interface{}),
	}

	if len(options) > 0 { //不管有几个,目前只有第一个有效
		r.Options = options[0]
	}

	return r
}

/* }}} */

/* {{{ func getRouteKey(rt *Route) (key string)
 *
 */
func getRouteKey(rt *Route) (key string) {
	key = fmt.Sprint(strings.ToUpper(rt.Method), " ", rt.Pattern)
	return
}

/* }}} */

/* {{{ func handlerWrap(rt *Route) web.HandlerFunc
 * 封装
 */
func handlerWrap(rt *Route) web.HandlerFunc { //这里封装了webC到本地的结构中
	fn := func(c web.C, w http.ResponseWriter, r *http.Request) {
		// build newest RESTContext
		rc := rcHolder(c, w, r)

		//route
		rc.Route = rt

		if nl, ok := rt.Options[NoLogKey]; ok && nl == true {
			rc.SetEnv(NoLogKey, true)
		}

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

/* }}} */

/* {{{ func (rtr *Router) New(mux *Mux, endpoint string)
 *
 */
func (rtr *Router) New(mux *Mux, endpoint string) {
	rtr.Mux = mux
	rtr.Endpoint = endpoint
}

/* }}} */

/* {{{ func (rtr *Router) GetEndpoint() string
 *
 */
func (rtr *Router) GetEndpoint() string {
	return rtr.Endpoint
}

/* }}} */

/* {{{ func (rtr *Router) Init(c RouterInterface)
 *
 */
func (rtr *Router) Init(c RouterInterface) {
	//rtr.Endpoint = endpoint
	rtr.DefaultRoutes(c) //默认路由
	if len(rtr.Routes) > 0 {
		for _, rt := range rtr.SRoutes {
			//Debug("pattern: %s", rt.Pattern)
			key := getRouteKey(rt)
			// regist routes to Mux
			rtr.Mux.Routes[key] = rt
			switch strings.ToLower(rt.Method) {
			case "get":
				rtr.RouteGet(rt)
			case "post":
				rtr.RoutePost(rt)
			case "put":
				rtr.RoutePut(rt)
			case "delete":
				rtr.RouteDelete(rt)
			case "patch":
				rtr.RoutePatch(rt)
			case "head":
				rtr.RouteHead(rt)
			default:
				// unknow method
			}
		}
	}
	// not found
	notFoundRoute := &Route{
		Handler: c.NotFound,
		Options: map[string]interface{}{NoLogKey: true},
	}
	rtr.RouteNotFound(notFoundRoute)
}

/* }}} */

/* {{{ Routers默认Action
 *
 */
func (rtr *Router) Get(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (rtr *Router) Post(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (rtr *Router) Put(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (rtr *Router) Delete(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (rtr *Router) Patch(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (rtr *Router) Head(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (rtr *Router) Options(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (rtr *Router) Trace(c *RESTContext) {
	c.HTTPError(http.StatusMethodNotAllowed)
}
func (rtr *Router) NotFound(c *RESTContext) {
	c.HTTPError(http.StatusNotFound)
}

/* }}} */

/* {{{ func (rtr *Router) AddRoute(m string, p interface{}, h Handler, options ...map[string]interface{})
 *
 */
func (rtr *Router) AddRoute(m string, p interface{}, h Handler, options ...map[string]interface{}) {
	key := fmt.Sprint(strings.ToUpper(m), " ", p)
	if rtr.Routes == nil {
		rtr.Routes = make(map[string]*Route)
		rtr.SRoutes = make([]*Route, 0)
	}
	if _, ok := rtr.Routes[key]; ok {
		//手动加路由, 如果冲突则以最早的为准
		Info("route dup: %s", key)
	} else {
		rtr.Routes[key] = NewRoute(p, m, h, options...)
		rtr.SRoutes = append(rtr.SRoutes, rtr.Routes[key])
	}
}

/* }}} */

/* {{{ func (rtr *Router) DefaultRoutes(c RouterInterface)
 * 默认路由, 如果已经定义了则忽略，没有定义则加上
 */
func (rtr *Router) DefaultRoutes(c RouterInterface) {
	if rtr.Endpoint == "" {
		//没有endpoint,不需要默认路由
		Info("Not need default Routes because no endpoint")
		return
	}
	var pattern, method, key string
	// GET /{endpoint}
	pattern = "/" + rtr.Endpoint
	method = "GET"
	key = method + " " + pattern
	if rtr.Routes == nil {
		rtr.Routes = make(map[string]*Route)
		rtr.SRoutes = make([]*Route, 0)
	}
	if _, ok := rtr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
		Warn("default route dup: %s", key)
	} else {
		rt := NewRoute(pattern, method, c.Get)
		rtr.Routes[key] = rt
		rtr.SRoutes = append(rtr.SRoutes, rt)
	}

	// GET /{endpoint}/{id}
	pattern = "/" + rtr.Endpoint + "/:_id_"
	method = "GET"
	key = method + " " + pattern
	if _, ok := rtr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
	} else {
		rt := NewRoute(pattern, method, c.Get)
		rtr.Routes[key] = rt
		rtr.SRoutes = append(rtr.SRoutes, rt)
	}

	// POST /{endpoint}
	pattern = "/" + rtr.Endpoint
	method = "POST"
	key = method + " " + pattern
	if _, ok := rtr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
	} else {
		rt := NewRoute(pattern, method, c.Post)
		rtr.Routes[key] = rt
		rtr.SRoutes = append(rtr.SRoutes, rt)
	}

	// DELETE /{endpoint}/{id}
	pattern = "/" + rtr.Endpoint + "/:_id_"
	method = "DELETE"
	key = method + " " + pattern
	if _, ok := rtr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
	} else {
		rt := NewRoute(pattern, method, c.Delete)
		rtr.Routes[key] = rt
		rtr.SRoutes = append(rtr.SRoutes, rt)
	}

	// PATCH /{endpoint}/{id}
	pattern = "/" + rtr.Endpoint + "/:_id_"
	method = "PATCH"
	key = method + " " + pattern
	if _, ok := rtr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
	} else {
		rt := NewRoute(pattern, method, c.Patch)
		rtr.Routes[key] = rt
		rtr.SRoutes = append(rtr.SRoutes, rt)
	}

	// PUT /{endpoint}/{id}
	pattern = "/" + rtr.Endpoint + "/:_id_"
	method = "PUT"
	key = method + " " + pattern
	if _, ok := rtr.Routes[key]; ok {
		// exists, warning, 默认路由不能覆盖自定义路由
	} else {
		rt := NewRoute(pattern, method, c.Put)
		rtr.Routes[key] = rt
		rtr.SRoutes = append(rtr.SRoutes, rt)
	}
}

/* }}} */

/* {{{ goji's methods
 *
 */
func (rtr *Router) RouteGet(rt *Route) {
	goji.Get(rt.Pattern, handlerWrap(rt))
}

func (rtr *Router) RoutePost(rt *Route) {
	goji.Post(rt.Pattern, handlerWrap(rt))
}

func (rtr *Router) RoutePut(rt *Route) {
	goji.Put(rt.Pattern, handlerWrap(rt))
}

func (rtr *Router) RouteDelete(rt *Route) {
	goji.Delete(rt.Pattern, handlerWrap(rt))
}

func (rtr *Router) RoutePatch(rt *Route) {
	goji.Patch(rt.Pattern, handlerWrap(rt))
}

func (rtr *Router) RouteHead(rt *Route) {
	goji.Head(rt.Pattern, handlerWrap(rt))
}

func (rtr *Router) RouteNotFound(rt *Route) {
	goji.NotFound(handlerWrap(rt))
}

/* }}} */

/* {{{ func CRUD(m Model, flag int) Handler
 * 通用的操作方法, 根据flag返回
 * 必须符合通用的restful风格
 */
func CRUD(m Model, flag int) Handler {
	get := func(c *RESTContext) {
	}
	post := func(c *RESTContext) {
	}
	delete := func(c *RESTContext) {
	}
	patch := func(c *RESTContext) { //修改
	}
	//put := func(c *RESTContext) { //重置
	//}
	head := func(c *RESTContext) { //检查字段
	}
	deny := func(c *RESTContext) {
	}

	switch flag {
	case GC_GET:
		return get
	case GC_POST:
		return post
	case GC_DELETE:
		return delete
	case GC_PATCH:
		return patch
	//case GC_PUT:
	//	return put
	case GC_HEAD:
		return head
	default:
		return deny
	}
}

/* }}} */
