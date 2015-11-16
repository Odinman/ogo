package ogo

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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
	EmptyGifBytes, _ = base64.StdEncoding.DecodeString(base64GifPixel)
)

// http context, 封装第三方包goji
type RESTContext struct {
	web.C
	Response      http.ResponseWriter
	Request       *http.Request
	Status        int
	ContentLength int
	RequestBody   []byte
	Accept        int //接受类型
	Version       string
	OTP           *OTPSpec
	Access        *Access
	Route         *Route
}

type OTPSpec struct {
	Value string
	Type  string
	Sn    string
}

type RESTError struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
	status  int
}

// implement error interface
func (re *RESTError) Error() string { return re.Message }

/* {{{ func newContext(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext
 *
 */
func newContext(c web.C, w http.ResponseWriter, r *http.Request) (*RESTContext, error) {
	rc := &RESTContext{
		C:        c,
		Response: w,
		Request:  r,
	}

	// default json
	rc.Accept = ContentTypeJSON

	//request body
	if r.Method != "GET" && r.Method != "HEAD" && r.Method != "DELETE" {
		//rc.Trace("content-type: %s", r.Header.Get("Content-Type"))
		if strings.Contains(r.Header.Get("Content-Type"), "multipart/") {
			rc.Trace("parse multipart")
			if err := r.ParseMultipartForm(env.MaxMemory); err != nil {
				rc.Error("parse multipart form error: %s", err)
				return rc, err
			}
		} else {
			rc.RequestBody, _ = ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			//ReadAll会清空r.Body, 下面需要写回去
			//bf := bytes.NewBuffer(rc.RequestBody)
			//r.Body = ioutil.NopCloser(bf)
		}
	}

	return rc, nil
}

/* }}} */

/* {{{ func rcHolder(c web.C, w http.ResponseWriter, r *http.Request) (func(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext)
 * 利用闭包初始化RESTContext, 并防止某些关键字段被重写(RequestBody)
 */
func RCHolder(c web.C, w http.ResponseWriter, r *http.Request) (*RESTContext, func(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext, error) {

	//初始化, RequestBody之类的保持住
	rc, err := newContext(c, w, r)

	if err != nil {
		return rc, nil, err
	}

	fn := func(c web.C, w http.ResponseWriter, r *http.Request) *RESTContext {
		rc.C = c
		rc.Response = w
		rc.Request = r
		return rc
	}

	return rc, fn, nil
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
		Message: message,
		Errors:  errors,
		status:  status,
	}
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTGenericError(status int, msg interface{}) (err error)
 * 普通错误,就是没有抓到error时报的错
 */
func (rc *RESTContext) RESTGenericError(status int, msg interface{}) (err error) {
	rc.SetStatus(status)
	// write data
	if msg != nil {
		err = rc.Output(rc.NewRESTError(status, msg))
	} else {
		err = rc.Output(nil)
	}
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTOK(data interface{}) (err error)
 * 属于request的错误
 */
func (rc *RESTContext) RESTOK(data interface{}) (err error) {
	if rc.Status <= 0 {
		if rc.OTP != nil {
			return rc.RESTTFA(data)
		} else {
			var status int
			method := strings.ToLower(rc.Request.Method)
			if _, ok := SUCCODE[method]; !ok {
				status = http.StatusOK //默认都是StatusOK
			} else {
				status = SUCCODE[method]
			}
			rc.SetStatus(status)
		}
	}

	// write data
	err = rc.Output(data)
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTTFA(data interface{}) (err error)
 * rest tfa回应
 */
func (rc *RESTContext) RESTTFA(data interface{}) (err error) {
	if rc.OTP != nil {
		rc.SetStatus(http.StatusAccepted)
		rc.SetHeader(otpHeader, fmt.Sprintf("%s; %s=%s", rc.OTP.Value, rc.OTP.Type, strconv.Quote(rc.OTP.Sn)))

		// write data
		err = rc.Output(data)
		return
	} else {
		return rc.RESTOK(data)
	}
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
	rc.SetStatus(status)

	// write data
	if msg != nil {
		err = rc.Output(rc.NewRESTError(status, msg))
	} else {
		err = rc.Output(nil)
	}
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
		rc.SetStatus(re.status)
		return rc.Output(re)
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
