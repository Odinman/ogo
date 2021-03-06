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
	EmptyGifBytes, _ = base64.StdEncoding.DecodeString(base64GifPixel)
)

type Task struct {
	Queue string
	Tag   string
	Value string
}

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
	App           interface{}
	tasks         []*Task
	locks         map[string]*Lock //访问锁
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

/* {{{ func getCode(ifSuc bool, m string) (s int)
 * 获取返回码
 */
func getCode(ifSuc bool, m string) (s int) {
	method := strings.ToLower(m)
	if ifSuc {
		switch method {
		case "get":
			return http.StatusOK
		case "delete":
			return http.StatusNoContent
		case "put":
			return http.StatusCreated
		case "post":
			return http.StatusCreated
		case "patch":
			return http.StatusResetContent
		case "head":
			return http.StatusOK
		default:
			return http.StatusOK
		}
	} else {
		switch method {
		case "get":
			return http.StatusNotFound
		case "delete":
			return http.StatusNotAcceptable
		case "put":
			return http.StatusNotAcceptable
		case "post":
			return http.StatusNotAcceptable
		case "patch":
			return http.StatusNotAcceptable
		case "head":
			return http.StatusConflict
		default:
			return http.StatusBadRequest
		}
	}
}

/* }}} */

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
	if rc.Access != nil && rc.Access.Session != "" {
		errors["session"] = rc.Access.Session
	}

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
			rc.SetStatus(getCode(true, rc.Request.Method))
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
	status := getCode(false, rc.Request.Method)
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

/* {{{ func (rc *RESTContext) GetLock(key string) (err error)
 *
 */
func (rc *RESTContext) GetLock(key string) (err error) {
	if rc.locks == nil {
		rc.locks = make(map[string]*Lock)
	}
	if _, ok := rc.locks[key]; !ok {
		rc.locks[key] = NewLock(key)
	}
	if err = rc.locks[key].Get(); err != nil {
		rc.Info("get lock(%s) error: %s", key, err)
	} else {
		rc.Debug("get lock(%s) ok, checksum: %s", key, rc.locks[key].checksum)
	}
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) PushTask(queue, tag, value string) error
 *
 */
func (rc *RESTContext) PushTask(queue, tag, value string) error {
	if queue == "" || tag == "" || value == "" {
		return fmt.Errorf("task_format_error")
	}
	if rc.tasks == nil {
		rc.tasks = make([]*Task, 0)
	}
	rc.tasks = append(rc.tasks, &Task{Queue: queue, Tag: tag, Value: value})
	return nil
}

/* }}} */

/* {{{ func (rc *RESTContext) LaunchTasks() error
 *
 */
func (rc *RESTContext) LaunchTasks() error {
	if rc.Status >= 400 || rc.tasks == nil || len(rc.tasks) <= 0 { // 400以内代表成功
		return fmt.Errorf("not_need_launch")
	}
	for _, t := range rc.tasks {
		if err := OmqTask(t.Queue, t.Tag, t.Value); err != nil {
			rc.Info("[queue: %s][tag: %s][value: %s][failed]", t.Queue, t.Tag, t.Value)
		} else {
			rc.Debug("[queue: %s][tag: %s][value: %s]", t.Queue, t.Tag, t.Value)
		}
	}
	return nil
}

/* }}} */

/* {{{ func (rc *RESTContext) ReleaseLocks()
 * 释放所有锁
 */
func (rc *RESTContext) ReleaseLocks() {
	if rc.locks == nil {
		return
	}
	for key, lk := range rc.locks {
		if err := lk.Release(); err != nil {
			rc.Info("release lock(%s) error: %s", key, err)
		} else {
			rc.Debug("release lock(%s) ok", key)
		}
	}
}

/* }}} */
