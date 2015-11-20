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

	// generic action const
	GA_GET = 1 << iota
	GA_SEARCH
	GA_POST
	GA_DELETE
	GA_PATCH
	//GA_PUT
	GA_HEAD
	GA_ALL = GA_GET | GA_SEARCH | GA_POST | GA_DELETE | GA_PATCH | GA_HEAD

	KEY_SKIPAUTH  = "skipauth"
	KEY_SKIPLOGIN = "skiplogin"
	KEY_SKIPPERM  = "skipperm"
	KEY_TPL       = "tpl"

	//env key
	RequestIDKey      = "_reqid_"
	SaveBodyKey       = "_sb_"
	NoLogKey          = "_nl_"
	PaginationKey     = "_pagination_"
	FieldsKey         = "_fields_"
	TimeRangeKey      = "_tr_"
	OrderByKey        = "_ob_"
	ConditionsKey     = "_conditions_"
	LogPrefixKey      = "_prefix_"
	EndpointKey       = "_endpoint_"
	RowkeyKey         = "_rk_"
	SelectorKey       = "_selector_"
	MimeTypeKey       = "_mimetype_"
	DispositionMTKey  = "_dmt_"
	ContentMD5Key     = "_md5_"
	DispositionPrefix = "_dp_"
	DIMENSION_KEY     = "_dimension_" //在restcontext中的key
	SIDE_KEY          = "_sidekey_"
	USERID_KEY        = "_userid_"
	APPID_KEY         = "_appid_"
	STAG_KEY          = "_stag_"
	PERMISSION_KEY    = "_perm_"
	EXT_KEY           = "_ext_"

	//1x1 gir
	base64GifPixel = "R0lGODlhAQABAIAAAP///wAAACwAAAAAAQABAAACAkQBADs="
	//base64GifPixel = "R0lGODlhAQABAJAAAP8AAAAAACH5BAUQAAAALAAAAAABAAEAAAICBAEAOw=="

	// 内容类型
	ContentTypeJSON = 1 << iota
	ContentTypeHTML
	ContentTypeXML
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
