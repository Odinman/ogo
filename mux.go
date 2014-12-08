// Ogo

package ogo

/* {{{ import
 */
import (
	"github.com/Odinman/ogo/libs/config"
	"github.com/Odinman/ogo/libs/logs"
)

/* }}} */

/* }}} */

/* {{{ type Mux struct
 */
type Mux struct {
	Env     *Environment           //环境参数
	Cfg     config.ConfigContainer //配置信息
	Logger  *logs.OLogger          //日志记录
	Workers map[string]*Worker
	Routes  map[string]*Route
	Hooks   HStack
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

/* {{{ func (mux *Mux) NewController(c ControllerInterface)
 * 这样做的目的是给controller设置mux(mux可多个) -- mux=multiplexer,复用器
 */
func (mux *Mux) NewController(c ControllerInterface) ControllerInterface {
	c.SetMux(mux)
	return c
}

/* }}} */
