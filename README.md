## ogo

Odin's go daemon/HTTP framework

## Original Features

* Daemonize, 可在<appname>.conf中用 Daemonize={bool}配置, pidfile默认写到程序目录的run/<appname>.pid
* DebugLevel配置,<appname>.conf中 DebugLevel={int}配置,数字越高级别越高
* Support HTTP! 如果在<appname>.conf中配置Service="HTTP"(不限大小写), 则可以编写webserver

## Daemon
* main.go如下:

```
package main

import (
    "github.com/zhaocloud/ogo"
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
    "github.com/zhaocloud/ogo"
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
	"github.com/zhaocloud/ogo"
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
	"github.com/zhaocloud/ogo"
	"time"
)

var UC *U

func init() {
	UC = &U{ogo.Controller{
		Endpoint: "user",
		Routes:   make(map[string]*ogo.Route),
	}}
	//载入默认路由
	//UC.DefaultRoutes(UC)
	UC.AddRoute("GET", "/odin/:_id_/:_selector_", UC.Test)
	UC.DefaultRoutes(UC)
}

type U struct {
	ogo.Controller
}

func (u *U) Get(ctx *ogo.HttpContext) {
	//ctx.Response.Write([]byte("123"))
	fmt.Fprintf(ctx.Response, "Hello, %s!", ctx.Context.URLParams["_id_"])
}

func (u *U) Post(ctx *ogo.HttpContext) {
	ctx.Response.Write([]byte("456"))
}

func (u *U) Test(ctx *ogo.HttpContext) {
	//ctx.Response.Write([]byte("123"))
	fmt.Fprintf(ctx.Response, "Hello, %s!, %s?", ctx.Context.URLParams["_id_"], ctx.Context.URLParams["_selector_"])
}
```

