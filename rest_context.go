package ogo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/zenazn/goji/web"
)

var (
	RESTC   *RESTContext
	SUCCODE = map[string]int{
		"get":    http.StatusOK,
		"delete": http.StatusNoContent,
		"put":    http.StatusCreated,
		"post":   http.StatusCreated,
		"patch":  http.StatusResetContent,
		"head":   http.StatusOK,
	}
	FAILCODE = map[string]int{ //定义正常出错
		"get":    http.StatusNotFound,
		"delete": http.StatusNotAcceptable,
		"put":    http.StatusNotAcceptable,
		"post":   http.StatusConflict, //冲突
		"patch":  http.StatusNotAcceptable,
		"head":   http.StatusConflict,
	}
)

// http context, 封装第三方包goji
type RESTContext struct {
	Context     web.C
	Response    http.ResponseWriter
	Request     *http.Request
	RequestBody []byte
}

type RESTError struct {
	Massage string            `json:"massage"`
	Errors  map[string]string `json:"errors"`
	status  int
}

// implement error interface
func (re *RESTError) Error() string { return re.Massage }

// new context
func newContext(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext {
	return &RESTContext{
		Context:  c,
		Response: w,
		Request:  r,
	}
}

func getContext(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext {
	if RESTC == nil {
		return newContext(c, w, r)
	} else {
		RESTC.Context = c //赋值,负责会被空覆盖
	}
	return RESTC
}

//new rest error
func (rc *RESTContext) NewRESTError(status int, msg interface{}) (re error) {
	errors := make(map[string]string)
	errors["method"] = rc.Request.Method
	errors["path"] = rc.Request.URL.Path
	errors["code"] = fmt.Sprint(status) // 备用, 可存储比httpstatus更详细的错误代码,目前只存httpstatus

	var message string
	if msg == nil {
		message = http.StatusText(status)
	} else {
		message = fmt.Sprint(msg)
	}
	re = &RESTError{
		Massage: message,
		Errors:  errors,
		status:  status,
	}
	return
}

// http error
func (rc *RESTContext) HTTPError(status int) (err error) {

	rc.RESTHeader(status)

	// write data
	err = rc.RESTBody(rc.NewRESTError(status, nil))

	return
}

func (rc *RESTContext) RESTHeader(status int) {
	// Content-Type always json
	rc.Response.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// header line
	rc.Response.WriteHeader(status)
}

func (rc *RESTContext) RESTBody(data interface{}) (err error) {

	var content []byte
	if Env.IndentJSON {
		content, _ = json.MarshalIndent(data, "", "  ")
	} else {
		content, _ = json.Marshal(data)
	}

	//write data
	_, err = rc.Response.Write(content)

	return
}

/* {{{ func (rc *RESTContext) RESTOK(data interface{}) (err error)
 * 属于request的错误
 */
func (rc *RESTContext) RESTOK(data interface{}) (err error) {
	var status int
	method := strings.ToLower(rc.Request.Method)
	if _, ok := SUCCODE[method]; !ok {
		status = http.StatusOK //默认都是StatusOK
	} else {
		status = SUCCODE[method]
	}
	rc.RESTHeader(status)

	// write data
	if data != nil {
		err = rc.RESTBody(data)
	}
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTNotOK(msg interface{}) (err error)
 * 属于request的错误
 */
func (rc *RESTContext) RESTNotOK(msg interface{}) (err error) {
	var status int
	method := strings.ToLower(rc.Request.Method)
	if _, ok := FAILCODE[method]; !ok {
		status = http.StatusOK //默认都是StatusOK
	} else {
		status = FAILCODE[method]
	}
	rc.RESTHeader(status)

	// write data
	if msg != nil {
		err = rc.RESTBody(rc.NewRESTError(status, msg))
	}
	return
}

/* }}} */

/* {{{ RESTGenericError
 * 普通错误,就是没有抓到error时报的错
 */
func (rc *RESTContext) RESTGenericError(status int, msg interface{}) (err error) {
	rc.RESTHeader(status)
	// write data
	err = rc.RESTBody(rc.NewRESTError(status, msg))
	return
}

/* }}} */

/* {{{ RESTNotFound
 *
 */
func (rc *RESTContext) RESTNotFound(msg interface{}) (err error) {
	return rc.RESTGenericError(http.StatusNotFound, msg)
}

/* }}} */

/* {{{ RESTPanic
 *
 */
func (rc *RESTContext) RESTPanic(msg interface{}) (err error) {
	return rc.RESTGenericError(http.StatusInternalServerError, msg)
}

/* }}} */

/* {{{ (rc *RESTContext) RESTBadRequest(msg interface{}) (err error)
 * BadRequest
 */
func (rc *RESTContext) RESTBadRequest(msg interface{}) (err error) {
	return rc.RESTGenericError(http.StatusBadRequest, msg)
}

/* }}} */

/* {{{ func (rc *RESTContext) GetQueryParam(key string) string
 */
func (rc *RESTContext) GetQueryParam(key string) string {
	v := rc.Request.Form[key]
	if len(v) == 1 {
		return string(v[0])
	} else {
		return string(strings.Join(v, ","))
	}
}

/* }}} */
