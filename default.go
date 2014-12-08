// Package ogo provides ...
package ogo

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
