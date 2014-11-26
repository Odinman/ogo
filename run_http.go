// +build !daemon

package ogo

/* {{{ import
 */
import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Odinman/ogo/bind"
	"github.com/Odinman/ogo/graceful"
	"github.com/VividCortex/godaemon"
	"github.com/nightlyone/lockfile"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web/middleware"
)

/* }}} */

/* {{{ func Run()
 * Run ogo application.
 */
func Run() {
	defer func() {
		if err := recover(); err != nil {
			WriteMsg("App crashed with error:", err)
			for i := 1; ; i++ {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				WriteMsg(file, line)
			}
			//panic要输出到console
			fmt.Println("App crashed with error:", err)
		}
	}()
	if Env.Daemonize {
		godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	}
	//check&write pidfile, added by odin
	dir := filepath.Dir(Env.PidFile)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			//mkdir
			if err := os.Mkdir(dir, 0755); err != nil {
				panic(err)
			}
		}
	}
	if l, err := lockfile.New(Env.PidFile); err == nil {
		if le := l.TryLock(); le != nil {
			panic(le)
		}
	} else {
		panic(err)
	}

	Debugger.Debug("will run http server")

	// 废除一些goji默认的middleware
	goji.Abandon(middleware.Logger)
	goji.Abandon(middleware.AutomaticOptions)

	//增加自定义的middleware
	goji.Use(EnvInit)
	goji.Use(Defer)
	goji.Use(Authentication)

	// in goji appengine mode (tags --appengine)
	goji.Serve()

	// socket listen
	bind.WithFlag()
	listener := bind.Default()
	Debugger.Warn("Starting Ogo on: %s", listener.Addr().String())

	graceful.HandleSignals()
	bind.Ready()
	graceful.PreHook(func() { Debugger.Warn("Received signal, gracefully stopping") })
	graceful.PostHook(func() { Debugger.Warn("Stopped") })

	err := graceful.Serve(listener, http.DefaultServeMux)

	if err != nil {
		Debugger.Critical(err.Error())
	}

	graceful.Wait()
}

/* }}} */
