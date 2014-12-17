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
	gojimiddle "github.com/zenazn/goji/web/middleware"
)

/* }}} */

func init() {
	// 废除全部goji默认的gojimiddle
	goji.Abandon(gojimiddle.RequestID)
	goji.Abandon(gojimiddle.Logger)
	goji.Abandon(gojimiddle.Recoverer)
	goji.Abandon(gojimiddle.AutomaticOptions)

	//增加自定义的middleware
	goji.Use(EnvInit)
	goji.Use(Defer)
	goji.Use(Mime)
	//goji.Use(Authentication)

}

/* {{{ func (mux *Mux) Run()
 * Run ogo application.
 */
func (mux *Mux) Run() {
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
	if env.Daemonize {
		godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	}
	//check&write pidfile, added by odin
	dir := filepath.Dir(env.PidFile)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			//mkdir
			if err := os.Mkdir(dir, 0755); err != nil {
				panic(err)
			}
		}
	}
	if l, err := lockfile.New(env.PidFile); err == nil {
		if le := l.TryLock(); le != nil {
			panic(le)
		}
	} else {
		panic(err)
	}

	Debug("will run http server")

	// in goji appengine mode (tags --appengine)
	goji.Serve()

	// socket listen
	bind.WithFlag()
	listener := bind.Default()
	Warn("Starting Ogo on: %s", listener.Addr().String())

	graceful.HandleSignals()
	bind.Ready()
	graceful.PreHook(func() { WriteMsg("Received signal, gracefully stopping") })
	graceful.PostHook(func() { WriteMsg("Stopped") })

	err := graceful.Serve(listener, http.DefaultServeMux)

	if err != nil {
		Critical(err.Error())
	}

	graceful.Wait()
}

/* }}} */
