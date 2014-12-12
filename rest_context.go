package ogo

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/zenazn/goji/web"
)

var (
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
		"post":   http.StatusNotAcceptable, //冲突
		"patch":  http.StatusNotAcceptable,
		"head":   http.StatusConflict,
	}
)

// http context, 封装第三方包goji
type RESTContext struct {
	web.C
	Response    http.ResponseWriter
	Request     *http.Request
	RequestBody []byte
	Route       *Route
}

type RESTError struct {
	Massage string            `json:"massage"`
	Errors  map[string]string `json:"errors"`
	status  int
}

// implement error interface
func (re *RESTError) Error() string { return re.Massage }

/* {{{ func newContext(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext
 *
 */
func newContext(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext {
	rc := &RESTContext{
		C:        c,
		Response: w,
		Request:  r,
	}

	//request body
	if r.Method != "GET" && r.Method != "HEAD" && r.Method != "DELETE" {
		rc.RequestBody, _ = ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		//ReadAll会清空r.Body, 下面需要写回去
		//bf := bytes.NewBuffer(rc.RequestBody)
		//r.Body = ioutil.NopCloser(bf)
	}
	//Debug("bodylen:%d", len(rc.RequestBody))

	// 解析参数
	r.ParseForm()

	return rc
}

/* }}} */

/* {{{ func rcHolder(c web.C, w http.ResponseWriter, r *http.Request) (func(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext)
 * 利用闭包初始化RESTContext, 并防止某些关键字段被重写(RequestBody)
 */
func RCHolder(c web.C, w http.ResponseWriter, r *http.Request) func(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext {

	//初始化, RequestBody之类的保持住
	rc := newContext(c, w, r)

	fn := func(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext {
		rc.C = c
		rc.Response = w
		rc.Request = r
		return rc
	}

	return fn
}

/* }}} */

/* {{{ func (rc *RESTContext) NewRESTError(status int, msg interface{}) (re error)
 *
 */
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

/* }}} */

/* {{{ func (rc *RESTContext) HTTPError(status int) (err error)
 *
 */
func (rc *RESTContext) HTTPError(status int) (err error) {

	rc.RESTHeader(status)

	// write data
	err = rc.RESTBody(rc.NewRESTError(status, nil))

	return
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTHeader(status int)
 *
 */
func (rc *RESTContext) RESTHeader(status int) {
	// Content-Type always json
	rc.Response.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// header line
	rc.Response.WriteHeader(status)
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTBody(data interface{}) (err error)
 *
 */
func (rc *RESTContext) RESTBody(data interface{}) (err error) {

	var content []byte
	if env.IndentJSON {
		content, _ = json.MarshalIndent(data, "", "  ")
	} else {
		content, _ = json.Marshal(data)
	}

	//write data
	_, err = rc.Response.Write(content)

	return
}

/* }}} */

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
		status = http.StatusBadRequest //默认都是StatusOK
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

/* {{{ func (rc *RESTContext) RESTGenericError(status int, msg interface{}) (err error)
 * 普通错误,就是没有抓到error时报的错
 */
func (rc *RESTContext) RESTGenericError(status int, msg interface{}) (err error) {
	rc.RESTHeader(status)
	// write data
	err = rc.RESTBody(rc.NewRESTError(status, msg))
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTNotFound(msg interface{}) (err error)
 *
 */
func (rc *RESTContext) RESTNotFound(msg interface{}) (err error) {
	return rc.RESTGenericError(http.StatusNotFound, msg)
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTPanic(msg interface{}) (err error)
 * 内部错误
 */
func (rc *RESTContext) RESTPanic(msg interface{}) (err error) {
	return rc.RESTGenericError(http.StatusInternalServerError, msg)
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTError(msg interface{}) (err error)
 * 内部错误
 */
func (rc *RESTContext) RESTError(err error) error {
	if re, ok := err.(*RESTError); ok {
		// 标准错误,直接输出
		rc.RESTHeader(re.status)
		return rc.RESTBody(re)
	} else {
		//普通错误, 普通输入
		rc.RESTNotOK(err)
	}
	return nil
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTBadRequest(msg interface{}) (err error)
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

/* {{{ func (rc *RESTContext) SetEnv(k string, v interface{})
 * 设置环境变量
 */
func (rc *RESTContext) SetEnv(k string, v interface{}) {
	if k != "" {
		rc.Env[k] = v
	}
}

/* }}} */

/* {{{ func (rc *RESTContext) GetEnv(k string) (v interface{})
 * 设置环境变量
 */
func (rc *RESTContext) GetEnv(k string) (v interface{}) {
	var ok bool
	if v, ok = rc.Env[k]; ok {
		return v
	}
	return nil
}

/* }}} */
