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
	Time      time.Time    `json:"t"`
	Status    int          `json:"sc"`
	Session   string       `json:"s"`
	IP        string       `json:"ip"`
	Method    string       `json:"m"`
	URI       string       `json:"uri"`
	Proto     string       `json:"p"`
	ReqBody   string       `json:"rbd,omitempty"`
	ReqLength int          `json:"ql"` //请求body大小
	RepLength int          `json:"pl"` //返回body大小
	Duration  string       `json:"d"`
	Host      string       `json:"h"`
	InHeader  *http.Header `json:"ih,omitempty"`
	OutHeader http.Header  `json:"oh,omitempty"`
	App       interface{}  `json:"app,omitempty"` //app自定义日志
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
