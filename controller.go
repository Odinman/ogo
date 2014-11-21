// Ogo

package ogo

import (
	//"fmt"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"net/http"
	"strings"
)

type Handler func(c *HttpContext)

type Route struct {
	Pattern string
	Method  string
	Handler Handler
}

type Controller struct {
	Endpoint string
	Routes   map[string]*Route
	Request  int
}

type ControllerInterface interface {
	Init(endpoint string)
	Get(c *HttpContext)
	Post(c *HttpContext)
	Put(c *HttpContext)
	Delete(c *HttpContext)
	Patch(c *HttpContext)
	Head(c *HttpContext)
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
		f(newContext(c, w, r))
	}
}

func (ctr *Controller) Init(endpoint string, c ControllerInterface) {
	ctr.Endpoint = endpoint
	ctr.Routes = make(map[string]*Route)
	//默认路由
	ctr.DefaultRoutes(c)
}

func (ctr *Controller) Get(c *HttpContext) {
	http.Error(c.Response, "Method Not Allowed", http.StatusMethodNotAllowed)
}
func (ctr *Controller) Post(c *HttpContext) {
	http.Error(c.Response, "Method Not Allowed", http.StatusMethodNotAllowed)
}
func (ctr *Controller) Put(c *HttpContext) {
	http.Error(c.Response, "Method Not Allowed", http.StatusMethodNotAllowed)
}
func (ctr *Controller) Delete(c *HttpContext) {
	http.Error(c.Response, "Method Not Allowed", http.StatusMethodNotAllowed)
}
func (ctr *Controller) Patch(c *HttpContext) {
	http.Error(c.Response, "Method Not Allowed", http.StatusMethodNotAllowed)
}
func (ctr *Controller) Head(c *HttpContext) {
	http.Error(c.Response, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (ctr *Controller) AddRoute(m string, p string, h Handler) {
	rt := NewRoute(p, m, h)
	switch strings.ToLower(m) {
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

// controller default route
// 默认路由, 如果已经定义了则忽略，没有定义则加上
//func (ctr *Controller) DefaultRoutes() {
func (ctr *Controller) DefaultRoutes(c ControllerInterface) {
	// GET /{endpoint}
	ctr.RouteGet(NewRoute("/"+ctr.Endpoint, "GET", c.Get))

	// GET /{endpoint}/{id}
	ctr.RouteGet(NewRoute("/"+ctr.Endpoint+"/:_id_", "GET", c.Get))

	// POST /{endpoint}
	ctr.RoutePost(NewRoute("/"+ctr.Endpoint, "POST", c.Post))

	// DELETE /{endpoint}/{id}
	ctr.RouteDelete(NewRoute("/"+ctr.Endpoint+"/:_id_", "DELETE", c.Delete))

	// PATCH /{endpoint}/{id}
	ctr.RouteDelete(NewRoute("/"+ctr.Endpoint+"/:_id_", "PATCH", c.Patch))

	// PUT /{endpoint}/{id}
	ctr.RouteDelete(NewRoute("/"+ctr.Endpoint+"/:_id_", "PUT", c.Put))

}

func (ctr *Controller) RouteGet(rt *Route) {
	key := strings.ToUpper(rt.Method) + " " + rt.Pattern
	if _, ok := ctr.Routes[key]; ok {
		// exists
	} else {
		goji.Get(rt.Pattern, handlerWrap(rt.Handler))
		ctr.Routes[key] = rt
	}
}
func (ctr *Controller) RoutePost(rt *Route) {
	key := strings.ToUpper(rt.Method) + " " + rt.Pattern
	if _, ok := ctr.Routes[key]; ok {
		// exists
	} else {
		goji.Post(rt.Pattern, handlerWrap(rt.Handler))
		ctr.Routes[key] = rt
	}
}
func (ctr *Controller) RoutePut(rt *Route) {
	key := strings.ToUpper(rt.Method) + " " + rt.Pattern
	if _, ok := ctr.Routes[key]; ok {
		// exists
	} else {
		goji.Put(rt.Pattern, handlerWrap(rt.Handler))
		ctr.Routes[key] = rt
	}
}
func (ctr *Controller) RouteDelete(rt *Route) {
	key := strings.ToUpper(rt.Method) + " " + rt.Pattern
	if _, ok := ctr.Routes[key]; ok {
		// exists
	} else {
		goji.Delete(rt.Pattern, handlerWrap(rt.Handler))
		ctr.Routes[key] = rt
	}
}
func (ctr *Controller) RoutePatch(rt *Route) {
	key := strings.ToUpper(rt.Method) + " " + rt.Pattern
	if _, ok := ctr.Routes[key]; ok {
		// exists
	} else {
		goji.Patch(rt.Pattern, handlerWrap(rt.Handler))
		ctr.Routes[key] = rt
	}
}
func (ctr *Controller) RouteHead(rt *Route) {
	key := strings.ToUpper(rt.Method) + " " + rt.Pattern
	if _, ok := ctr.Routes[key]; ok {
		// exists
	} else {
		goji.Head(rt.Pattern, handlerWrap(rt.Handler))
		ctr.Routes[key] = rt
	}
}
