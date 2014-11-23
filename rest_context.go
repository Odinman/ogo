package ogo

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zenazn/goji/web"
)

// http context, 封装第三方包goji
type RESTContext struct {
	Context    web.C
	Response   http.ResponseWriter
	Request    *http.Request
	Controller *Controller
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

//new rest error
func (hc *RESTContext) NewRESTError(status int, msg interface{}) (re error) {
	errors := make(map[string]string)
	errors["method"] = hc.Request.Method
	errors["path"] = hc.Request.URL.Path
	errors["code"] = fmt.Sprint(status) // 备用, 可存储比httpstatus更详细的错误代码,目前只存httpstatus

	var message string
	message = fmt.Sprint(msg)
	if message == "" {
		message = http.StatusText(status)
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

	// custom header
	hc.Response.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// header
	hc.Response.WriteHeader(status)

	// write data
	content, _ := json.MarshalIndent(hc.NewRESTError(status, nil), "", "  ")
	hc.Response.Write(content)

	return
}
