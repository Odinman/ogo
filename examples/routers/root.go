// zhaoonline storage service
package routers

import (
	"github.com/Odinman/ogo"
)

type RootRouter struct {
	ogo.Router
}

func init() {
	r := ogo.NewRouter(new(RootRouter), "").(*RootRouter) //无endpoint
	//先自定义路由, 因为优先级高, 自定义路由后定义的会覆盖先定义的
	r.AddRoute("GET", "/", r.Get, ogo.RouteOption{ogo.KEY_SKIPLOGIN: true, ogo.KEY_SKIPAUTH: true})
	r.AddRoute("GET", "/favicon.ico", r.Get, ogo.RouteOption{ogo.KEY_SKIPLOGIN: true, ogo.KEY_SKIPAUTH: true})
	r.AddRoute("POST", "/", r.Post, ogo.RouteOption{ogo.KEY_SKIPLOGIN: true, ogo.KEY_SKIPAUTH: true})
	r.AddRoute("DELETE", "/", r.Delete, ogo.RouteOption{ogo.KEY_SKIPLOGIN: true, ogo.KEY_SKIPAUTH: true})
	r.AddRoute("PATCH", "/", r.Patch, ogo.RouteOption{ogo.KEY_SKIPLOGIN: true, ogo.KEY_SKIPAUTH: true})
	// 初始化并载入默认路由, 默认路由不会覆盖自定义路由
	r.Init(r)
}

func (this *RootRouter) Root(c *ogo.RESTContext) {
}
