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

	_ "./hooks"
	_ "./routers"
)

func main() {
	ogo.Run()
}
```

* routers, 该目录下, 文件名可以自定, 范例如下:

```
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
    // do something
}
```

