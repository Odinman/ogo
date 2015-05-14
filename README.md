## ogo

Odin's go daemon/HTTP framework

## Original Features

* Daemonize, 可在<appname>.conf中用 Daemonize={bool}配置, pidfile默认写到程序目录的run/<appname>.pid
* DebugLevel配置,<appname>.conf中 DebugLevel={int}配置,数字越高级别越高
* Support HTTP! 默认支持http, 编译时候如果开启`-a --tags 'daemon'`, 则是daemon模式

## Daemon
* main.go如下:

```
package main

import (
    "github.com/Odinman/ogo"
    _ "<gopath_to_workers>/workers"
)

func main() {
    ogo.Run()
}
```

* workers目录下,文件名自定,以'test'为例,代码如下:

```
package workers

import (
    "github.com/Odinman/ogo"
)

type TestWorker struct {
    ogo.Worker
}

func init() {
    ogo.AddWorker(&TestWorker{})
}

func (w *TestWorker) Main() error {
    return nil
}

```

主要看worker的名字, 框架会自动调用Main()


## HTTP

* main.go

```
package main

import (
	"github.com/Odinman/ogo"
	_ "<gopath_to_controllers>/controller"
)

func main() {
	ogo.Run()
}
```

* controllers, 该目录下, 文件名可以自定, 范例如下:

```
package controller

import (
	"fmt"
	"github.com/Odinman/ogo"
)

var UC *U = &U{ogo.Controller{Routes: make(map[string]*ogo.Route)}} //先初始化路由

func init() {
	//先自定义路由, 因为优先级高, 自定义路由后定义的会覆盖先定义的
	UC.AddRoute("GET", "/user/:_rk_", UC.Test1)
	UC.AddRoute("GET", "/user/:_rk_", UC.Test2)
	// 初始化并载入默认路由, 默认路由不会覆盖自定义路由
	UC.Init("user", UC)

}

type U struct {
	ogo.Controller
}

func (u *U) Get(ctx *ogo.RESTContext) {
	//ctx.Response.Write([]byte("123"))
	u.Request += 1
	fmt.Fprintf(ctx.Response, "Hello, this is Get, %s!, %d??", ctx.Context.URLParams["_rk_"], u.Request)
}

func (u *U) Post(ctx *ogo.RESTContext) {
	ctx.Response.Write([]byte("456"))
}

func (u *U) Test1(ctx *ogo.RESTContext) {
	//ctx.Response.Write([]byte("123"))
	fmt.Fprintf(ctx.Response, "Test1, %s!, %s?", ctx.Context.URLParams["_rk_"], ctx.Context.URLParams["_selector_"])
}

func (u *U) Test2(ctx *ogo.RESTContext) {
	u.Request += 1
	//ctx.Response.Write([]byte("223"))
	fmt.Fprintf(ctx.Response, "Test2, %s!, %s?, %d!!", ctx.Context.URLParams["_rk_"], ctx.Context.URLParams["_selector_"], u.Request)
}
```

