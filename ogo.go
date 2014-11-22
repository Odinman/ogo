// Ogo

package ogo

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"time"

	"github.com/Odinman/ogo/libs/config"
	"github.com/Odinman/ogo/libs/logs"
	"github.com/VividCortex/godaemon"
	"github.com/nightlyone/lockfile"
	"github.com/zenazn/goji"
	//"github.com/zenazn/goji/web/middleware"
)

// ogo daemoin framework version.
const VERSION = "0.1.0"

type Context struct {
	Env     *Environment           //环境参数
	Cfg     config.ConfigContainer //配置信息
	Workers map[string]*Worker
	Logger  *logs.OLogger //日志记录
}

var (
	Ctx       *Context
	Env       *Environment
	AppConfig config.ConfigContainer
	Debugger  *logs.OLogger
)

func NewContext() *Context {
	return &Context{
		Workers: make(map[string]*Worker),
	}
}

// Run ogo application.
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

	var mainErr error

	if Env.Service != "http" { //正常daemon
		Debugger.Debug("will run worker: %v", Env.Worker)
		if worker, ok := Ctx.Workers[Env.Worker]; ok {
			vw := reflect.New(worker.WorkerType)
			execWorker, ok := vw.Interface().(WorkerInterface)
			if !ok {
				panic("worker is not WorkerInterface")
			}

			//Init
			execWorker.Init(Ctx, Env.Worker)

			//Main
			mainErr = execWorker.Main()
		} else {
			mainErr = errors.New("not found worker: " + Env.Worker)
		}

		if mainErr != nil {
			panic(mainErr)
		}
	} else {

		// 废除一些goji默认的middleware
		//goji.Abandon(middleware.Logger)
		//goji.Abandon(middleware.AutomaticOptions)

		//增加自定义的middleware
		goji.Use(func(h http.Handler) http.Handler {
			handler := func(w http.ResponseWriter, r *http.Request) {
				log.Println("Before request")
				h.ServeHTTP(w, r)
				log.Println("After request")
			}
			return http.HandlerFunc(handler)
		})
		goji.Use(func(h http.Handler) http.Handler {
			handler := func(w http.ResponseWriter, r *http.Request) {
				log.Println("Before request 111")
				h.ServeHTTP(w, r)
				log.Println("After request 111")
			}
			return http.HandlerFunc(handler)
		})

		goji.Serve()
	}
	//睡一段时间再结束
	time.Sleep(1000 * time.Microsecond)
}
