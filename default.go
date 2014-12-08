// Package ogo provides ...
package ogo

import (
	"github.com/Odinman/ogo/libs/config"
	"github.com/Odinman/ogo/libs/logs"
)

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

/* {{{ func Config() config.ConfigContainer
 *
 */
func Config() config.ConfigContainer {
	return DMux.Config()
}

/* }}} */

/* {{{ func Logger() config.LoggerContainer
 *
 */
func Logger() *logs.OLogger {
	return DMux.Logger()
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
