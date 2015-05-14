// Package ogo provides ...
package ogo

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Odinman/ogo/libs/config"
	"github.com/Odinman/ogo/libs/logs"
)

/* {{{ func init()
 *
 */
func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	DMux = New() // default mux

	if err := DMux.initEnv(); err != nil {
		// SetEnv包含了环境以及配置的初始化, logger也放里面
		fmt.Printf("init env failed: %s", err)
		os.Exit(1)
	}
}

/* }}} */

/* {{{ func Run()
 * 默认Run()
 */
func Run() {
	DMux.Run()
}

/* }}} */

/* {{{ func NewRouter(c RouterInterface, endpoint string) RouterInterface
 * 默认用DMux设置给router
 */
func NewRouter(c RouterInterface, endpoint string) RouterInterface {
	return DMux.NewRouter(c, endpoint)
}

/* }}} */

/* {{{ func Env() *Environ
 * 默认的配置就是DMux的配置
 */
func Env() *Environ {
	if env, err := DMux.Env(); err != nil {
		return nil
	} else {
		return env
	}
}

/* }}} */

/* {{{ func Config() config.ConfigContainer
 * 默认的配置就是DMux的配置
 */
func Config() config.ConfigContainer {
	if cfg, err := DMux.Config(); err != nil {
		return nil
	} else {
		return cfg
	}
}

/* }}} */

/* {{{ func Logger() *logs.OLogger
 *
 */
func Logger() *logs.OLogger {
	if logger, err := DMux.Logger(); err != nil {
		return nil
	} else {
		return logger
	}
}

/* }}} */

/* {{{ func Accessor() *logs.OLogger
 *
 */
func Accessor() *logs.OLogger {
	if accessor, err := DMux.Accessor(); err != nil {
		return nil
	} else {
		return accessor
	}
}

/* }}} */

/* {{{ func PreHook(hook OgoHook)
 * 正式程序之前的钩子
 */
func PreHook(hook OgoHook) {
	DMux.PreHook(hook)
}

/* }}} */

/* {{{ func PostHook(hook OgoHook)
 * 正式程序之前的钩子
 */
func PostHook(hook OgoHook) {
	DMux.PostHook(hook)
}

/* }}} */

/* {{{ func AddTagHook(tag string, hook TagHook)
 * tag hook
 */
func AddTagHook(tag string, hook TagHook) {
	DMux.AddTagHook(tag, hook)
}

/* }}} */
