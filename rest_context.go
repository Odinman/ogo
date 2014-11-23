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
	}
	return RESTC
}

//new rest error
func (hc *RESTContext) NewRESTError(status int, msg interface{}) (re error) {
	errors := make(map[string]string)
	errors["method"] = hc.Request.Method
	errors["path"] = hc.Request.URL.Path
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
func (hc *RESTContext) HTTPError(status int) (err error) {

	hc.RESTHeader(status)

	// write data
	err = hc.RESTBody(hc.NewRESTError(status, nil))

	return
}

func (hc *RESTContext) RESTHeader(status int) {
	// Content-Type always json
	hc.Response.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// header line
	hc.Response.WriteHeader(status)
}

func (hc *RESTContext) RESTBody(data interface{}) (err error) {

	content, _ := json.MarshalIndent(data, "", "  ")

	//write data
	_, err = hc.Response.Write(content)

	return
}
