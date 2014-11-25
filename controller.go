// Ogo

package ogo

import (
	//"fmt"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"net/http"
	"strings"
)

type Handler func(c *RESTContext)

type Route struct {
	Pattern string
	Method  string
	Handler Handler
}

type Controller struct {
	Endpoint string
	Routes   map[string]*Route
	ReqCount int //访问计数
}

type ControllerInterface interface {
	//Init(endpoint string, c ControllerInterface)
	Get(c *RESTContext)
	Post(c *RESTContext)
	Put(c *RESTContext)
	Delete(c *RESTContext)
	Patch(c *RESTContext)
	Head(c *RESTContext)
	Options(c *RESTContext)
	Trace(c *RESTContext)
	NotFound(c *RESTContext)
}

func NewRoute(p string, m string, h Handler) *Route {
	return &Route{
		Pattern: p,
		Method:  m,
		Handler: h,
	}
}

// 封装
func handlerWrap(f Handler) web.HandlerFunc { //这里封装了webC到本地的结构中
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		f(getContext(c, w, r))
	}
}

func (ctr *Controller) Init(endpoint string, c ControllerInterface) {
	ctr.Endpoint = endpoint
	//ctr.Routes = make(map[string]*Route)
	//默认路由
	ctr.DefaultRoutes(c)
	if len(ctr.Routes) > 0 {
		for _, rt := range ctr.Routes {
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
	ctr.RouteNotFound(c.NotFound)
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

func (ctr *Controller) AddRoute(m string, p string, h Handler) {
	key := strings.ToUpper(m) + " " + p
	if ctr.Routes == nil {
		ctr.Routes = make(map[string]*Route)
	}
	if _, ok := ctr.Routes[key]; ok {
		//手动加路由, 以最后加的为准,overwrite
	}
	ctr.Routes[key] = NewRoute(p, m, h)
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
	goji.Get(rt.Pattern, handlerWrap(rt.Handler))
}

func (ctr *Controller) RoutePost(rt *Route) {
	goji.Post(rt.Pattern, handlerWrap(rt.Handler))
}

func (ctr *Controller) RoutePut(rt *Route) {
	goji.Put(rt.Pattern, handlerWrap(rt.Handler))
}

func (ctr *Controller) RouteDelete(rt *Route) {
	goji.Delete(rt.Pattern, handlerWrap(rt.Handler))
}

func (ctr *Controller) RoutePatch(rt *Route) {
	goji.Patch(rt.Pattern, handlerWrap(rt.Handler))
}

func (ctr *Controller) RouteHead(rt *Route) {
	goji.Head(rt.Pattern, handlerWrap(rt.Handler))
}

func (ctr *Controller) RouteNotFound(h Handler) {
	goji.NotFound(handlerWrap(h))
}
