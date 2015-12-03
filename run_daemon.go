// +build daemon

package ogo

/* {{{ import
 */
import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"time"

	"github.com/VividCortex/godaemon"
	"github.com/nightlyone/lockfile"
	//"github.com/zenazn/goji"
	//"github.com/zenazn/goji/web/middleware"
)

var (
	stdOut io.Reader
)

const (
	stageKey = "__OGO_STAGE"
)

/* }}} */

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
	//if env.Daemonize {
	//	godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	//}
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

	var mainErr error

	Debug("will run worker: %v", env.Worker)
	if worker, ok := DMux.Workers[env.Worker]; ok {
		vw := reflect.New(worker.WorkerType)
		execWorker, ok := vw.Interface().(WorkerInterface)
		if !ok {
			panic("worker is not WorkerInterface")
		}

		//Init
		execWorker.Init(DMux, env.Worker)

		//Main
		mainErr = execWorker.Main()
	} else {
		mainErr = fmt.Errorf("not found worker: %s", env.Worker)
	}

	if mainErr != nil {
		panic(mainErr)
	}
	//睡一段时间再结束
	time.Sleep(1000 * time.Microsecond)
}

/* }}} */
