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
	Time    time.Time `json:"t"`
	Session string    `json:"s"`
	IP      string    `json:"ip"`
	Method  string    `json:"m"`
	URI     string    `json:"uri"`
	Proto   string    `json:"p"`
	Status  int       `json:"sc"`
	Size    int       `json:"sz"`
	//Duration  time.Duration `json:"d"`
	Duration  string       `json:"d"`
	Host      string       `json:"h"`
	InHeader  *http.Header `json:"ih,omitempty"`
	OutHeader http.Header  `json:"oh,omitempty"`
	App       interface{}  `json:"app,omitempty"` //app自定义日志
}

/* {{{ func (ac *Access) Save()
 *
 */
func (ac *Access) Save() {
	if ab, err := json.Marshal(ac); err == nil {
		accessor.Access(string(ab))
	}
}

/* }}} */
