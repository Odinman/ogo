// Debugger

package ogo

import (
	"encoding/json"
	"net/http"
	"time"
	//"strings"

	//"github.com/Odinman/ogo/libs/logs"
)

type Access struct {
	Time     time.Time   `json:"t"`
	Service  string      `json:"sn,omitempty"`
	Session  string      `json:"s"`
	Duration string      `json:"d"`
	Http     *HTTPLog    `json:"http,omitempty"`
	App      interface{} `json:"app,omitempty"` //app自定义日志
}

type HTTPLog struct {
	Status    int          `json:"sc"`
	IP        string       `json:"ip"`
	Method    string       `json:"m"`
	URI       string       `json:"uri"`
	Proto     string       `json:"p"`
	ReqBody   string       `json:"rbd,omitempty"`
	ReqLength int          `json:"ql"` //请求body大小
	RepLength int          `json:"pl"` //返回body大小
	Host      string       `json:"h"`
	InHeader  *http.Header `json:"ih,omitempty"`
	OutHeader http.Header  `json:"oh,omitempty"`
}

/* {{{ func (ac *Access) Save()
 * 记录access日志
 */
func (ac *Access) Save() {
	if ab, err := json.Marshal(ac); err == nil {
		accessor.Access(string(ab))
	}
}

/* }}} */
