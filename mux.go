// Ogo

package ogo

/* {{{ import
 */
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Odinman/ogo/libs/config"
	"github.com/Odinman/ogo/libs/logs"
	"github.com/Odinman/ogo/utils"
)

/* }}} */

/* {{{ type Environ struct
 */
type Environ struct {
	lock          *sync.RWMutex
	WorkPath      string         // working path(abs)
	AppPath       string         // application path
	ProcName      string         // proc name
	Worker        string         // worker name
	AppConfigPath string         // config file path
	RunMode       string         // run mode, "dev" or "prod"
	AccessPath    string         //acces log file path
	Daemonize     bool           // daemonize or not
	EnableGzip    bool           // enable gzip or not
	DebugLevel    int            // debug level
	PidFile       string         // pidfile abs path
	Port          string         // http port
	IndentJSON    bool           // indent JSON
	MaxMemory     int64          //max memory(form-data)
	Location      *time.Location // location
	initErr       error
}

/* }}} */

/* {{{ type Mux struct
 */
type Mux struct {
	env      *Environ               //环境参数
	cfg      config.ConfigContainer //配置信息
	logger   *logs.OLogger          //debug日志记录
	accessor *logs.OLogger          //日志
	Workers  map[string]*Worker
	Routes   map[string]*Route
	Hooks    HStack
}

/* }}} */

/* {{{ func New() *Mux
 */
func New() *Mux {
	return &Mux{
		Workers: make(map[string]*Worker),
		Routes:  make(map[string]*Route),
		Hooks: HStack{
			preHooks:  make([]OgoHook, 0),
			postHooks: make([]OgoHook, 0),
		},
	}
}

/* }}} */

/* {{{ func (mux *Mux) PreHook(hook OgoHook)
 * 正式程序之前的钩子
 */
func (mux *Mux) PreHook(hook OgoHook) {
	mux.Hooks.preHooks = append(mux.Hooks.preHooks, hook)
}

/* }}} */

/* {{{ func (mux *Mux) PostHook(hook OgoHook)
 * 正式程序之后的钩子
 */
func (mux *Mux) PostHook(hook OgoHook) {
	mux.Hooks.postHooks = append(mux.Hooks.postHooks, hook)
}

/* }}} */

/* {{{ func (mux *Mux) NewRouter(c RouterInterface, endpoint string) RouterInterface
 * 这样做的目的是给rounter设置mux(mux可多个) -- mux=multiplexer,复用器
 */
func (mux *Mux) NewRouter(c RouterInterface, endpoint string) RouterInterface {
	c.New(mux, endpoint)
	return c
}

/* }}} */

/* {{{ func (mux *Mux) Env(path string) *Environ, error
 * 获取env变量
 */
func (mux *Mux) Env() (*Environ, error) {

	if mux.env == nil {
		env = &Environ{
			lock: new(sync.RWMutex),
		}

		// default value
		env.RunMode = "dev"     //default runmod
		env.Port = DEFAULT_PORT //default is 8001
		env.Daemonize = false
		env.DebugLevel = logs.LevelTrace //默认debug等级
		env.IndentJSON = false
		env.MaxMemory = 1 << 26                              //64MB
		env.Location, _ = time.LoadLocation("Asia/Shanghai") //默认上海时区

		workPath, _ := os.Getwd()
		env.WorkPath, _ = filepath.Abs(workPath)
		env.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		env.ProcName = filepath.Base(os.Args[0])   //程序名字
		env.Worker = strings.ToLower(env.ProcName) //worker默认为procname,小写
		env.AccessPath = "logs/access.log"         //默认access日志是程序目录下的logs/access.log

		//默认配置文件是 conf/{ProcName}.conf
		env.AppConfigPath = filepath.Join(env.AppPath, "conf", env.ProcName+".conf")
		if !utils.FileExists(env.AppConfigPath) {
			//不存在时指定为app.conf
			tp := filepath.Join(env.AppPath, "conf", "app.conf")
			if utils.FileExists(tp) {
				env.AppConfigPath = tp
			}
		}

		// pidfile depend AppPath
		env.PidFile = filepath.Join(env.AppPath, "run", env.ProcName+".pid")

		if env.WorkPath != env.AppPath {
			//切换工作目录
			os.Chdir(env.AppPath)
		}

		//配置文件必须存在
		if !utils.FileExists(env.AppConfigPath) {
			return nil, fmt.Errorf("config file not found: %s", env.AppConfigPath)
		}

		mux.env = env
	}

	return mux.env, nil
}

/* }}} */

/* {{{ func (mux *Mux) initEnv(cfg config.ConfigContainer) error
 * 设置环境变量
 */
func (mux *Mux) initEnv() (err error) {
	if env, err = mux.Env(); err != nil {
		return err
	}
	if cfg, err = mux.Config(); err != nil {
		return err
	}

	//根据配置设置环境变量
	env.lock.Lock()
	defer env.lock.Unlock()

	// location
	if tz := cfg.String("TimeZone"); tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			env.Location = loc
		}
	}

	if port := cfg.String("Port"); port != "" {
		env.Port = port
	}
	os.Setenv("PORT", env.Port) // pass to bind

	if runmode := cfg.String("RunMode"); runmode != "" {
		env.RunMode = runmode
	}

	if workerName := cfg.String("Worker"); workerName != "" {
		env.Worker = workerName
	}

	if daemonize, err := cfg.Bool("Daemonize"); err == nil {
		env.Daemonize = daemonize
	}
	if enablegzip, err := cfg.Bool("EnableGzip"); err == nil {
		env.EnableGzip = enablegzip
	}
	if indentJson, err := cfg.Bool("IndentJson"); err == nil {
		env.IndentJSON = indentJson
	}
	// 自定义pidfile
	if pidfile := cfg.String("PidFile"); pidfile != "" {
		// make sure pidfile is abs path
		if filepath.IsAbs(pidfile) {
			env.PidFile = pidfile
		} else {
			env.PidFile = filepath.Join(env.AppPath, pidfile)
		}
	}
	//自定义access 日志位置
	if apath := cfg.String("AccessPath"); apath != "" {
		env.AccessPath = apath
	}
	if level, err := cfg.Int("DebugLevel"); err == nil {
		env.DebugLevel = level
	}

	// logger init
	if logger, err = mux.Logger(); err != nil {
		return err
	}

	//access init
	if accessor, err = mux.Accessor(); err != nil {
		return err
	}

	return nil
}

/* }}} */

/* {{{ func (mux *Mux) Config() (config.ConfigContainer, error)
 * 获取配置信息
 */
func (mux *Mux) Config() (config.ConfigContainer, error) {
	if mux.cfg == nil {
		if env, err := mux.Env(); err != nil {
			return nil, err
		} else {
			cp := env.AppConfigPath
			if cfg, err := config.NewConfig("ini", cp); err != nil {
				return nil, err
			} else {
				mux.cfg = cfg
			}
		}
	}
	return mux.cfg, nil
}

/* }}} */

/* {{{ func (mux *Mux) Logger() config.LoggerContainer
 *
 */
func (mux *Mux) Logger() (*logs.OLogger, error) {
	if mux.logger == nil {
		// init logger
		logger := logs.NewLogger(2046)
		var err error
		if mux.env.Daemonize {
			err = logger.SetLogger("file", `{"filename":"logs/debug.log"}`)
		} else {
			err = logger.SetLogger("console", "")
		}

		if err != nil {
			return nil, err
		}

		logger.EnableFuncCallDepth(true)
		logger.SetLogFuncCallDepth(4)
		logger.SetLevel(mux.env.DebugLevel)
		mux.logger = logger
	}

	return mux.logger, nil
}

/* }}} */

/* {{{ func (mux *Mux) Accessor() config.LoggerContainer
 *
 */
func (mux *Mux) Accessor() (*logs.OLogger, error) {
	if mux.accessor == nil {
		// init logger
		logger := logs.NewLogger(2046)
		var err error
		err = logger.SetLogger("access", fmt.Sprintf(`{"filename":"%s"}`, mux.env.AccessPath))

		if err != nil {
			return nil, err
		}

		mux.accessor = logger
	}

	return mux.accessor, nil
}

/* }}} */
