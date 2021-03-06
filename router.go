// Ogo

package ogo

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Odinman/ogo/utils"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

type Handler func(c *RESTContext)

type RouteOption map[interface{}]interface{}

type Route struct {
	Key      string //独立标识
	Endpoint string
	Pattern  interface{}
	Method   string
	Handler  Handler
	Options  *utils.SafeMap
	Updating bool
	Creating bool
}

type Router struct {
	Endpoint   string
	Routes     map[string]*Route
	Hooks      map[string]TagHook
	SRoutes    []*Route //排序的Route
	ReqCount   int      //访问计数
	Mux        *Mux
	Controller interface{} //既是RouterInterface, 也是 ActionInterface
}

type RouterInterface interface {
	New(interface{}, *Mux, string)
	AddRoute(m string, p interface{}, h Handler, options ...RouteOption)
	DefaultRoutes() //默认路由
	GetEndpoint() string
	Init()

	//method
	Get(*RESTContext)
	Post(*RESTContext)
	Put(*RESTContext)
	Delete(*RESTContext)
	Patch(*RESTContext)
	Head(*RESTContext)
	Options(*RESTContext)
	Trace(*RESTContext)
	NotFound(*RESTContext)

	// action
	CRUD(i interface{}, flag int) Handler
}

/* {{{ func NewRoute(p interface{}, ep string, m string, h Handler, options ...RouteOption) *Route
 *
 */
func NewRoute(p interface{}, ep string, m string, h Handler, options ...RouteOption) *Route {
	r := &Route{
		Key:      fmt.Sprint(strings.ToUpper(m), " ", p),
		Endpoint: ep,
		Pattern:  p,
		Method:   m,
		Handler:  h,
	}

	if len(options) > 0 { //不管有几个,目前只有第一个有效
		r.Options = utils.NewSafeMap(options[0])
	} else {
		r.Options = utils.NewSafeMap()
	}

	//更新还是创建
	if m == "POST" {
		r.Creating = true
	} else if m == "PATCH" {
		r.Updating = true
	}

	return r
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

		if nl := rt.Options.Get(NoLogKey); nl != nil && nl.(bool) == true {
			rc.SetEnv(NoLogKey, true)
		}

		// pre hooks, 任何一个出错,都要结束
		if hl := len(DMux.Hooks.preHooks); hl > 0 {
			DMux.Hooks.lock.Lock()
			defer DMux.Hooks.lock.Unlock()
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
			DMux.Hooks.lock.Lock()
			defer DMux.Hooks.lock.Unlock()
			for i := 0; i < hl; i++ {
				DMux.Hooks.postHooks[i](rc)
			}
		}
	}
	return fn
}

/* }}} */

/* {{{ func (rtr *Router) New(c interface{},mux *Mux, endpoint string)
 *
 */
func (rtr *Router) New(c interface{}, mux *Mux, endpoint string) {
	rtr.Controller = c
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

/* {{{ func (rtr *Router) Init()
 *
 */
func (rtr *Router) Init() {
	ri := rtr.Controller.(RouterInterface)
	//rtr.Endpoint = endpoint
	rtr.DefaultRoutes() //默认路由
	if len(rtr.Routes) > 0 {
		for _, rt := range rtr.SRoutes {
			//Debug("pattern: %s", rt.Pattern)
			key := rt.Key
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
		Handler: ri.NotFound,
		Options: utils.NewSafeMap(map[interface{}]interface{}{NoLogKey: true}),
	}
	rtr.RouteNotFound(notFoundRoute)
}

/* }}} */

/* {{{ Routers Default Action
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
func (rtr *Router) EmptyHtml(c *RESTContext) {
	c.HTTPOK(nil)
}
func (rtr *Router) EmptyGif(c *RESTContext) {
	c.HTTPEmptyGif()
}

/* }}} */

/* {{{ func (rtr *Router) AddRoute(m string, p interface{}, h Handler, options ...RouteOption)
 *
 */
func (rtr *Router) AddRoute(m string, p interface{}, h Handler, options ...RouteOption) {
	if rtr.Routes == nil {
		rtr.Routes = make(map[string]*Route)
		rtr.SRoutes = make([]*Route, 0)
	}
	rt := NewRoute(p, rtr.GetEndpoint(), m, h, options...)
	key := rt.Key
	if _, ok := rtr.Routes[key]; ok {
		//手动加路由, 如果冲突则以最早的为准
		Info("route dup: %s", key)
	} else {
		rtr.Routes[key] = rt
		rtr.SRoutes = append(rtr.SRoutes, rtr.Routes[key])
	}
}

/* }}} */

/* {{{ func (rtr *Router) DefaultRoutes()
 * 默认路由, 如果已经定义了则忽略，没有定义则加上
 */
func (rtr *Router) DefaultRoutes() {
	ri := rtr.Controller.(RouterInterface)
	if rtr.Endpoint == "" {
		//没有endpoint,不需要默认路由
		Info("Not need default Routes because no endpoint")
		return
	}

	var method, pattern string

	// HEAD /{endpoint}
	method = "HEAD"
	pattern = "/" + rtr.Endpoint
	rtr.AddRoute(method, pattern, ri.Head)

	// HEAD /{endpoint}/{id}
	method = "HEAD"
	pattern = "/" + rtr.Endpoint + "/:" + RowkeyKey
	rtr.AddRoute(method, pattern, ri.Head)

	// GET /{endpoint}
	pattern = "/" + rtr.Endpoint
	method = "GET"
	rtr.AddRoute(method, pattern, ri.Get)

	// GET /{endpoint}/{id}
	pattern = "/" + rtr.Endpoint + "/:" + RowkeyKey
	method = "GET"
	rtr.AddRoute(method, pattern, ri.Get)

	// POST /{endpoint}
	pattern = "/" + rtr.Endpoint
	method = "POST"
	rtr.AddRoute(method, pattern, ri.Post)

	// DELETE /{endpoint}/{id}
	pattern = "/" + rtr.Endpoint + "/:" + RowkeyKey
	method = "DELETE"
	rtr.AddRoute(method, pattern, ri.Delete)

	// PATCH /{endpoint}/{id}
	pattern = "/" + rtr.Endpoint + "/:" + RowkeyKey
	method = "PATCH"
	rtr.AddRoute(method, pattern, ri.Patch)

	// PUT /{endpoint}/{id}
	pattern = "/" + rtr.Endpoint + "/:" + RowkeyKey
	method = "PUT"
	rtr.AddRoute(method, pattern, ri.Put)
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

/* {{{ func (rtr *Router) GenericRoute(i interface{}, flag int)
 * 自动路由, 任何implelent了Action的类型都可以使用
 */
func (rtr *Router) GenericRoute(i interface{}, flag int) {
	endpoint := rtr.GetEndpoint()
	if flag&GA_HEAD > 0 {
		// HEAD /{endpoint}
		rtr.AddRoute("HEAD", "/"+endpoint, rtr.CRUD(i, GA_HEAD), RouteOption{KEY_SKIPLOGIN: true}) //HEAD默认无需登录
	}
	if flag&GA_GET > 0 {
		// GET /{endpoint}
		rtr.AddRoute("GET", "/"+endpoint, rtr.CRUD(i, GA_SEARCH))
		// GET /{endpoint}/{id}
		rtr.AddRoute("GET", "/"+endpoint+"/:"+RowkeyKey, rtr.CRUD(i, GA_GET))
	}
	if flag&GA_POST > 0 {
		// POST /{endpoint}
		rtr.AddRoute("POST", "/"+endpoint, rtr.CRUD(i, GA_POST))
	}
	if flag&GA_DELETE > 0 {
		// DELETE /{endpoint}/{id}
		rtr.AddRoute("DELETE", "/"+endpoint+"/:"+RowkeyKey, rtr.CRUD(i, GA_DELETE))
	}
	if flag&GA_PATCH > 0 {
		// PATCH /{endpoint}/{id}
		rtr.AddRoute("PATCH", "/"+endpoint+"/:"+RowkeyKey, rtr.CRUD(i, GA_PATCH))
	}
	//if flag&GA_PUT > 0 {
	//	// PUT /{endpoint}/{id}
	//	rtr.AddRoute("PUT", "/"+endpoint+"/:"+RowkeyKey, CRUD(m, GA_PUT))
	//}
}

/* }}} */

/* {{{ func (rtr *Router) CRUD(i interface{}, flag int) Handler
 * 通用的操作方法, 根据flag返回
 * 必须符合通用的restful风格
 */
func (rtr *Router) CRUD(i interface{}, flag int) Handler {
	act := rtr.Controller.(ActionInterface)
	get := func(c *RESTContext) {
		m := NewModel(i.(Model), c)
		defer act.Defer(m)

		if _, err := act.PreGet(m); err != nil {
			c.Warn("PreGet error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		var r interface{}
		var err error
		if r, err = act.OnGet(m); err != nil {
			c.Warn("OnGet error: %s", err)
			if err == ErrNoRecord {
				c.RESTNotFound(err)
			} else {
				c.RESTPanic(err)
			}
			return
		}

		if r, err = act.PostGet(r); err != nil {
			c.Warn("PostGet error: %s", err)
			c.RESTNotOK(err)
		} else {
			c.RESTOK(r)
		}

		return
	}
	search := func(c *RESTContext) {
		m := NewModel(i.(Model), c)
		defer act.Defer(m)

		if _, err := act.PreSearch(m); err != nil { // presearch准备条件等
			c.Warn("PreSearch error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		if l, err := act.OnSearch(m); err != nil {
			c.Warn("OnSearch error: %s", err)
			if err == ErrNoRecord {
				c.RESTNotFound(err)
			} else {
				c.RESTPanic(err)
			}
		} else {
			if rl, err := act.PostSearch(l); err != nil {
				c.Warn("PostSearch error: %s", err)
				c.RESTNotOK(err)
			} else {
				c.RESTOK(rl)
			}
		}

		return

	}

	post := func(c *RESTContext) {
		m := NewModel(i.(Model), c)
		defer act.Defer(m)
		var err error

		if _, err = act.PreCreate(m); err != nil { // presearch准备条件等
			c.Warn("PreCreate error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		var r interface{}
		if r, err = act.OnCreate(m); err != nil {
			c.Warn("OnCreate error: %s", err)
			c.RESTNotOK(err)
			return
		}
		m = r.(Model)

		// 触发器
		r, err = act.Trigger(m)
		if err != nil {
			c.Warn("Trigger error: %s", err)
		}

		// create ok, return
		if r, err = act.PostCreate(r); err != nil {
			c.Warn("postCreate error: %s", err)
		}
		c.AppLoggingResult(r)
		c.RESTOK(r)
		return
	}

	delete := func(c *RESTContext) {
		m := NewModel(i.(Model), c)
		defer act.Defer(m)
		var err error

		if _, err = act.PreDelete(m); err != nil { // presearch准备条件等
			c.Warn("PreUpdat error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		if _, err = act.OnDelete(m); err != nil {
			c.Warn("OnUpdate error: %s", err)
			c.RESTNotOK(err)
			return
		}

		// update ok
		var r interface{}
		if r, err = act.PostDelete(m); err != nil {
			c.Warn("postCreate error: %s", err)
		}

		// 触发器
		_, err = act.Trigger(m)
		if err != nil {
			c.Warn("Trigger error: %s", err)
		}
		c.RESTOK(r)
		return
	}

	patch := func(c *RESTContext) { //修改
		m := NewModel(i.(Model), c)
		defer act.Defer(m)
		var err error

		if _, err = act.PreUpdate(m); err != nil { // presearch准备条件等
			c.Warn("PreUpdate error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		if _, err = act.OnUpdate(m); err != nil {
			c.Warn("OnUpdate error: %s", err)
			c.RESTNotOK(err)
			return
		}

		// 触发器
		_, err = act.Trigger(m)
		if err != nil {
			c.Warn("Trigger error: %s", err)
		}

		// update ok
		var r interface{}
		if r, err = act.PostUpdate(m); err != nil {
			c.Warn("postCreate error: %s", err)
		}

		c.AppLoggingResult(r)
		c.RESTOK(r)
		return
	}
	//put := func(c *RESTContext) { //重置
	//}
	head := func(c *RESTContext) { //检查字段
		m := NewModel(i.(Model), c)
		defer act.Defer(m)

		if _, err := act.PreCheck(m); err != nil { // presearch准备条件等
			c.Warn("PreCheck error: %s", err)
			c.RESTBadRequest(err)
			return
		}

		if cnt, err := act.OnCheck(m); err != nil {
			c.Warn("OnCheck error: %s", err)
			if err == ErrNoRecord {
				c.RESTNotFound(err)
			} else {
				c.RESTPanic(err)
			}
		} else {
			if cnt, _ := act.PostCheck(cnt); cnt.(int64) > 0 {
				c.RESTNotOK(nil)
			} else {
				c.RESTOK(nil)
			}
		}

		return
	}
	deny := func(c *RESTContext) {
		c.HTTPError(http.StatusMethodNotAllowed)
	}

	switch flag {
	case GA_GET:
		return get
	case GA_SEARCH:
		return search
	case GA_POST:
		return post
	case GA_DELETE:
		return delete
	case GA_PATCH:
		return patch
	//case GA_PUT:
	//	return put
	case GA_HEAD:
		return head
	default:
		return deny
	}
}

/* }}} */
