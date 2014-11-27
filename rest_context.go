package ogo

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zenazn/goji/web"
)

var RESTC *RESTContext

// http context, 封装第三方包goji
type RESTContext struct {
	Context  web.C
	Response http.ResponseWriter
	Request  *http.Request
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

// rest not found
func (rc *RESTContext) RESTNotFound(msg interface{}) (err error) {
	rc.RESTHeader(http.StatusNotFound)

	// write data
	err = rc.RESTBody(rc.NewRESTError(http.StatusNotFound, msg))
	return
}

// rest ok
func (rc *RESTContext) RESTOK(data interface{}) (err error) {
	rc.RESTHeader(http.StatusOK)

	// write data
	err = rc.RESTBody(data)
	return
}
