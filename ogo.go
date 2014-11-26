// Ogo

package ogo

/* {{{ import
 */
import (
	"github.com/Odinman/ogo/libs/config"
	"github.com/Odinman/ogo/libs/logs"
)

/* }}} */

/* {{{ const
 */
const (
	// ogo daemoin framework version.
	VERSION = "0.1.0"
)

/* }}} */

/* {{{ type Context struct
 */
type Context struct {
	Env     *Environment           //环境参数
	Cfg     config.ConfigContainer //配置信息
	Workers map[string]*Worker
	Logger  *logs.OLogger //日志记录
}

/* }}} */

/* {{{ variables
 */
var (
	Ctx       *Context
	Env       *Environment
	AppConfig config.ConfigContainer
	Debugger  *logs.OLogger
)

/* }}} */

/* {{{ func NewContext() *Context
 */
func NewContext() *Context {
	return &Context{
		Workers: make(map[string]*Worker),
	}
}

/* }}} */
