// zhaoonline storage service
package routers

import (
	"github.com/Odinman/ogo"

	"../models"
)

type TestRouter struct {
	ogo.Router
}

func init() {
	r := ogo.NewRouter(new(TestRouter), "test").(*TestRouter)
	//先自定义路由, 因为优先级高, 自定义路由后定义的会覆盖先定义的
	//r.AddRoute("GET", "/test", r.Root, ogo.RouteOption{ogo.KEY_SKIPLOGIN: true, ogo.KEY_SKIPAUTH: true})
	// 初始化并载入默认路由, 默认路由不会覆盖自定义路由
	r.GenericRoute(new(models.Test), ogo.GA_ALL)
	r.Init(r)
}

func (this *TestRouter) Root(c *ogo.RESTContext) {
}
