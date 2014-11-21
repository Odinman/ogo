package ogo

import (
	"net/http"

	"github.com/zenazn/goji/web"
)

// http context, 封装第三方包的好方法
type HttpContext struct {
	Context  web.C
	Response http.ResponseWriter
	Request  *http.Request
}

func newContext(c web.C, w http.ResponseWriter, r *http.Request) *HttpContext {
	return &HttpContext{
		Context:  c,
		Response: w,
		Request:  r,
	}
}
