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
	// ogo daemon/http framework version.
	VERSION      = "0.1.0"
	DEFAULT_PORT = "8001"
)

/* }}} */

/* {{{ variables
 */
var (
	DMux   *Mux
	env    *Environment
	cfg    config.ConfigContainer
	logger *logs.OLogger
)

/* }}} */
