// Ogo

package ogo

/* {{{ import
 */
import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Odinman/ogo/libs/config"
	"github.com/Odinman/ogo/libs/logs"
	"github.com/Odinman/ogo/utils"
)

/* }}} */

/* {{{ const
 */
const (
	// ogo daemon/http framework version.
	VERSION = "0.1.0"
)

/* }}} */

/* {{{ variables
 */
var (
	DMux      *Mux
	env       *Environment
	appConfig config.ConfigContainer
	debugger  *logs.OLogger
)

/* }}} */

/* {{{ func init()
 *
 */
func init() {
	DMux = New() // default mux

	env = new(Environment)
	env.Port = "8001" //default is 8001
	workPath, _ := os.Getwd()
	env.WorkPath, _ = filepath.Abs(workPath)
	env.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	env.ProcName = filepath.Base(os.Args[0])   //程序名字
	env.Worker = strings.ToLower(env.ProcName) //worker默认为procname,小写
	env.IndentJSON = false

	//默认配置文件是 conf/{ProcName}.conf
	env.AppConfigPath = filepath.Join(env.AppPath, "conf", env.ProcName+".conf")
	if !utils.FileExists(env.AppConfigPath) {
		//不存在时指定为app.conf
		env.AppConfigPath = filepath.Join(env.AppPath, "conf", "app.conf")
	}

	if env.WorkPath != env.AppPath {
		if utils.FileExists(env.AppConfigPath) {
			os.Chdir(env.AppPath)
		} else {
			//在当前目录找配置文件
			env.AppConfigPath = filepath.Join(env.WorkPath, "conf", env.ProcName+".conf")
			if !utils.FileExists(env.AppConfigPath) {
				//不存在时指定为app.conf
				env.AppConfigPath = filepath.Join(env.WorkPath, "conf", "app.conf")
			}
		}
	}

	env.RunMode = "dev" //default runmod

	env.Daemonize = false
	env.PidFile = filepath.Join(env.AppPath, "run", env.ProcName+".pid")
	env.DebugLevel = logs.LevelTrace //默认debug等级

	runtime.GOMAXPROCS(runtime.NumCPU())

	configErr := ParseConfig()

	DMux.Env = env

	// init debugger
	debugger = logs.NewLogger(2046)
	var err error
	if env.Daemonize {
		//fmt.Println("file")
		err = debugger.SetLogger("file", `{"filename":"logs/debug.log"}`)
	} else {
		//fmt.Println("console")
		err = debugger.SetLogger("console", "")
	}
	if err != nil {
		fmt.Println("init logger error:", err)
	}
	debugger.EnableFuncCallDepth(true)
	debugger.SetLogFuncCallDepth(2)
	debugger.SetLevel(env.DebugLevel)
	//Debug("hihi")

	if configErr != nil {
		//放在这里才能使用Logger函数
		Log("Warn", "%v", configErr)
	}

	DMux.Cfg = appConfig
	DMux.Logger = debugger
}

/* }}} */
