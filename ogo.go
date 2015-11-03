// Ogo

package ogo

/* {{{ import
 */
import (
	"github.com/Odinman/ogo/libs/config"
	"github.com/Odinman/ogo/libs/logs"
	omq "github.com/Odinman/omq/utils"
)

/* }}} */

/* {{{ const
 */
const (
	// ogo daemon/http framework version.
	VERSION      = "1.0"
	DEFAULT_PORT = "8001"
)

/* }}} */

/* {{{ variables
 */
var (
	DMux     *Mux
	env      *Environ
	cfg      config.ConfigContainer
	logger   *logs.OLogger
	accessor *logs.OLogger
	omqpool  *omq.Pool
)

/* }}} */
