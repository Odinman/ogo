// +build !daemon

package ogo

/* {{{ import
 */
import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
	goji.Use(ParseParams)

	//mime
	initMime()
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

	Warn("Starting Ogo on: %s", env.Port)
}

/* }}} */
