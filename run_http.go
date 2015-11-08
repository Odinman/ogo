// +build !daemon

package ogo

/* {{{ import
 */
import (
	"fmt"
	//"net/http"
	"bufio"
	"io"
	"os"
	"path/filepath"
	"runtime"

	//"github.com/Odinman/ogo/bind"
	//"github.com/Odinman/ogo/graceful"
	"github.com/VividCortex/godaemon"
	"github.com/nightlyone/lockfile"
	"github.com/zenazn/goji"
	gojimiddle "github.com/zenazn/goji/web/middleware"
)

/* }}} */

const (
	stageKey = "__OGO_STAGE"
)

var (
	stdOut io.Reader
)

func init() {
	// 废除全部goji默认的gojimiddle
	goji.Abandon(gojimiddle.RequestID)
	goji.Abandon(gojimiddle.Logger)
	goji.Abandon(gojimiddle.Recoverer)
	goji.Abandon(gojimiddle.AutomaticOptions)

	//增加自定义的middleware
	goji.Use(EnvInit)
	goji.Use(Defer)
	//goji.Use(Mime)
	goji.Use(ParseHeaders)
	goji.Use(ParseParams)

	//mime
	initMime()

	// root router
	rr := NewRouter(new(Router), "").(*Router)
	rr.AddRoute("GET", "/", rr.Get, RouteOption{KEY_SKIPLOGIN: true, KEY_SKIPAUTH: true})
	rr.AddRoute("GET", "/favicon.ico", rr.EmptyGif, RouteOption{KEY_SKIPLOGIN: true, KEY_SKIPAUTH: true, NoLogKey: true})
	rr.AddRoute("GET", "/index.html", rr.EmptyHtml, RouteOption{KEY_SKIPLOGIN: true, KEY_SKIPAUTH: true, NoLogKey: true})
	rr.AddRoute("POST", "/", rr.Post, RouteOption{KEY_SKIPLOGIN: true, KEY_SKIPAUTH: true})
	rr.AddRoute("DELETE", "/", rr.Delete, RouteOption{KEY_SKIPLOGIN: true, KEY_SKIPAUTH: true})
	rr.AddRoute("PATCH", "/", rr.Patch, RouteOption{KEY_SKIPLOGIN: true, KEY_SKIPAUTH: true})
	rr.Init()
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
		//  for debug, CaptureOutput
		stdOut, _, _ = godaemon.MakeDaemon(&godaemon.DaemonAttr{CaptureOutput: true})
		go func(reader io.Reader) {
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				Debug("[stdOut] %s", scanner.Text())
			}
		}(stdOut)
	} else {
		if processStage := os.Getenv(stageKey); processStage == "" { //头一次
			Debug("processStage: %s", processStage)
			os.Setenv(stageKey, "ogo")
			if procName, err := godaemon.GetExecutablePath(); err != nil || len(procName) == 0 {
				panic(err)
			} else {
				files := make([]*os.File, 3)
				files[0], files[1], files[2] = os.Stdin, os.Stdout, os.Stderr
				dir, _ := os.Getwd()
				osAttrs := os.ProcAttr{Dir: dir, Env: os.Environ(), Files: files}
				if proc, err := os.StartProcess(procName, os.Args, &osAttrs); err != nil {
					panic(err)
				} else {
					proc.Release()
					os.Exit(0)
				}
			}
		}
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
}

/* }}} */
