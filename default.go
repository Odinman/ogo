// Package ogo provides ...
package ogo

import (
	"fmt"
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

	if err := DMux.InitEnv(); err != nil {
		// SetEnv包含了环境以及配置的初始化, logger也放里面
		fmt.Printf("init env failed: %s", err)
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

/* {{{ func NewController()
 * 默认用DMux设置给controller
 */
func NewController(c ControllerInterface) ControllerInterface {
	return DMux.NewController(c)
}

/* }}} */

/* {{{ func Env() *Environment
 * 默认的配置就是DMux的配置
 */
func Env() *Environment {
	return env
}

/* }}} */

/* {{{ func Config() config.ConfigContainer
 * 默认的配置就是DMux的配置
 */
func Config() config.ConfigContainer {
	return cfg
}

/* }}} */

/* {{{ func Logger() *logs.OLogger
 *
 */
func Logger() *logs.OLogger {
	return logger
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
