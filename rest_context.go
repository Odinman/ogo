package ogo

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/zenazn/goji/web"
)

const (
	//env key
	RequestIDKey      = "_reqid_"
	SaveBodyKey       = "_sb_"
	SaveResultKey     = "_sr_"
	NoLogKey          = "_nl_"
	PaginationKey     = "_pagination_"
	FieldsKey         = "_fields_"
	TimeRangeKey      = "_tr_"
	OrderByKey        = "_ob_"
	ConditionsKey     = "_conditions_"
	LogPrefixKey      = "_prefix_"
	EndpointKey       = "_endpoint_"
	RowkeyKey         = "_rk_"
	SelectorKey       = "_selector_"
	MimeTypeKey       = "_mimetype_"
	DispositionMTKey  = "_dmt_"
	ContentMD5Key     = "_md5_"
	DispositionPrefix = "_dp_"

	DIMENSION_KEY  = "_dimension_" //在restcontext中的key
	SIDE_KEY       = "_sidekey_"
	USERID_KEY     = "_userid_"
	APPID_KEY      = "_appid_"
	PERMISSION_KEY = "_perm_"
	EXT_KEY        = "_ext_"

	//1x1 gir
	base64GifPixel = "R0lGODlhAQABAIAAAP///wAAACwAAAAAAQABAAACAkQBADs="
	//base64GifPixel = "R0lGODlhAQABAJAAAP8AAAAAACH5BAUQAAAALAAAAAABAAEAAAICBAEAOw=="
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
	Access        *Access
	Route         *Route
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

/* {{{ func (rc *RESTContext) WriteBytes(data []byte) (n int, e error)
 * 输出内容,如果需要压缩,统一在这里进行
 */
func (rc *RESTContext) WriteBytes(data []byte) (n int, e error) {
	if dLen := len(data); dLen > 0 { //有内容才需要
		if env.EnableGzip == true && rc.Request.Header.Get("Accept-Encoding") != "" {
			splitted := strings.SplitN(rc.Request.Header.Get("Accept-Encoding"), ",", -1)
			encodings := make([]string, len(splitted))

			for i, val := range splitted {
				encodings[i] = strings.TrimSpace(val)
			}
			for _, val := range encodings {
				if val == "gzip" {
					rc.Response.Header().Set("Content-Encoding", "gzip")
					b := new(bytes.Buffer)
					w, _ := gzip.NewWriterLevel(b, gzip.BestSpeed)
					w.Write(data)
					w.Close()
					data = b.Bytes()
					break
				} else if val == "deflate" {
					rc.Response.Header().Set("Content-Encoding", "deflate")
					b := new(bytes.Buffer)
					w, _ := flate.NewWriter(b, flate.BestSpeed)
					w.Write(data)
					w.Close()
					data = b.Bytes()
					break
				}
			}
		}
		rc.ContentLength = dLen
		rc.Response.Header().Set("Content-Length", strconv.Itoa(rc.ContentLength))
	}
	if rc.Status == 0 {
		rc.Status = http.StatusOK
	}
	//在Write之前要WriteHeader
	rc.Response.WriteHeader(rc.Status)
	if len(data) > 0 {
		_, e = rc.Response.Write(data)
	}

	return
}

/* }}} */

/* {{{ func (rc *RESTContext) ServeBinary(mimetype string, data []byte)
 * 直接出二进制内容
 */
func (rc *RESTContext) ServeBinary(mimetype string, data []byte) {
	rc.Response.Header().Set("Content-Type", mimetype)
	rc.WriteBytes(data)
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTHeader(status int)
 *
 */
func (rc *RESTContext) RESTHeader(status int) {
	// Content-Type always json
	rc.Response.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// status
	//rc.Response.WriteHeader(status)
	rc.Status = status
}

/* }}} */

/* {{{ func (rc *RESTContext) RESTBody(data interface{}) (err error)
 *
 */
func (rc *RESTContext) RESTBody(data interface{}) (err error) {

	var content []byte
	if method := strings.ToLower(rc.Request.Method); method != "head" {
		if data != nil {
			if env.IndentJSON {
				content, _ = json.MarshalIndent(data, "", "  ")
			} else {
				content, _ = json.Marshal(data)
			}
		}

	}
	//write header & data
	_, err = rc.WriteBytes(content)

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

	if sr := rc.GetEnv(SaveResultKey); sr != nil && sr.(bool) == true {
		rc.Access.App = data
	}
	// write data
	err = rc.RESTBody(data)
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) HTTPOK(data []byte) (err error)
 * 属于request的错误
 */
func (rc *RESTContext) HTTPOK(data []byte) (err error) {
	rc.Response.Header().Set("Content-Type", "text/html; charset=UTF-8")
	rc.Status = http.StatusOK

	// write data
	_, err = rc.WriteBytes(data)
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) HTTPEmptyGif() (err error)
 * 属于request的错误
 */
func (rc *RESTContext) HTTPEmptyGif() (err error) {
	rc.Response.Header().Set("Content-Type", "image/gif")
	rc.Status = http.StatusOK

	// write data
	_, err = rc.WriteBytes(EmptyGifBytes)
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) HTTPBack() (err error)
 * 属于request的错误
 */
func (rc *RESTContext) HTTPBack() (err error) {
	rc.Status = http.StatusOK
	rc.Response.Header().Set("Content-Type", "text/html; charset=UTF-8")
	rc.Response.Header().Set("Cache-Control", "max-age=0")
	rc.Response.Header().Set("Cache-Control", "no-cache")
	rc.Response.Header().Set("Cache-Control", "must-revalidate")
	rc.Response.Header().Set("Cache-Control", "private")
	rc.Response.Header().Set("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")
	rc.Response.Header().Set("Pragma", "no-cache")

	// write data
	data := []byte(`<?xml version="1.0"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<meta http-equiv="Cache-Control" content="max-age=0" forua="true" />
<meta http-equiv="Cache-Control" content="no-cache" forua="true" />
<meta http-equiv="Cache-Control" content="must-revalidate" forua="true" />
<title></title>
</head>
<body><p><a href="javascript:history.back(1)">Back</a></p></body>
</html>`)
	_, err = rc.WriteBytes(data)
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) HTTPRedirect(url string) (err error)
 * 属于request的错误
 */
func (rc *RESTContext) HTTPRedirect(url string) (err error) {
	rc.Status = http.StatusFound //302
	rc.Response.Header().Set("Content-Type", "text/html; charset=UTF-8")
	rc.Response.Header().Set("Cache-Control", "max-age=0")
	rc.Response.Header().Set("Cache-Control", "no-cache")
	rc.Response.Header().Set("Cache-Control", "must-revalidate")
	rc.Response.Header().Set("Cache-Control", "private")
	rc.Response.Header().Set("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")
	rc.Response.Header().Set("Pragma", "no-cache")
	rc.Response.Header().Set("Location", url)

	err = rc.RESTBody(nil)
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
	} else {
		err = rc.RESTBody(nil)
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
	if msg != nil {
		err = rc.RESTBody(rc.NewRESTError(status, msg))
	} else {
		err = rc.RESTBody(nil)
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

/* {{{ func (rc *RESTContext) GetQueryParam(key string) (string, int)
 */
func (rc *RESTContext) GetQueryParam(key string) (r string, c int) {
	v := rc.Request.Form[key]
	c = len(v)
	if c == 1 {
		return string(v[0]), c
	} else {
		return string(strings.Join(v, ",")), c
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

/* {{{ func (rc *RESTContext) SaveAccess()
 * 设置环境变量
 */
func (rc *RESTContext) SaveAccess() {
	if nl := rc.GetEnv(NoLogKey); nl == true {
		return
	}
	if rc.Access != nil {
		rc.Access.Save()
	}
}

/* }}} */

/* {{{	RESTContext loggers
 * 可以在每个debug信息带上session
 */
func (rc *RESTContext) Trace(format string, v ...interface{}) {
	rc.logf("trace", format, v...)
}
func (rc *RESTContext) Debug(format string, v ...interface{}) {
	rc.logf("debug", format, v...)
}
func (rc *RESTContext) Info(format string, v ...interface{}) {
	rc.logf("info", format, v...)
}
func (rc *RESTContext) Print(format string, v ...interface{}) {
	rc.logf("info", format, v...)
}
func (rc *RESTContext) Warn(format string, v ...interface{}) {
	rc.logf("warn", format, v...)
}
func (rc *RESTContext) Error(format string, v ...interface{}) {
	rc.logf("error", format, v...)
}
func (rc *RESTContext) Critical(format string, v ...interface{}) {
	rc.logf("critical", format, v...)
}
func (rc *RESTContext) logf(tag, format string, v ...interface{}) {
	if nl := rc.GetEnv(NoLogKey); nl == true {
		// no logging
		return
	}
	var prefix string
	if p := rc.GetEnv(LogPrefixKey); p != nil {
		prefix = p.(string)
	}
	if prefix != "" {
		format = prefix + " " + format
	}
	switch strings.ToLower(tag) {
	case "trace":
		logger.Trace(format, v...)
	case "debug":
		logger.Debug(format, v...)
	case "info":
		logger.Info(format, v...)
	case "warn":
		logger.Warn(format, v...)
	case "error":
		logger.Error(format, v...)
	case "critial":
		logger.Critical(format, v...)
	default:
		logger.Debug(format, v...)
	}
}

/* }}} */
