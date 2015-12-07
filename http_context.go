package ogo

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Odinman/ogo/utils"
)

/* {{{ func (rc *RESTContext) SetHeader(k,v string)
 * set header
 */
func (rc *RESTContext) SetHeader(k, v string) {
	rc.Response.Header().Set(k, v)
}

/* }}} */

/* {{{ func (rc *RESTContext) SetStatus(status int)
 * 构建Status Code
 */
func (rc *RESTContext) SetStatus(status int) {
	//rc.Response.WriteHeader(status)
	rc.Status = status
}

/* }}} */

/* {{{ func (rc *RESTContext) Output(data interface{}) (err error)
 *
 */
func (rc *RESTContext) Output(data interface{}) (err error) {

	if rc.Status >= 200 && rc.Status < 300 && rc.Accept == ContentTypeHTML { //用户需要HTML
		tplFile := ""
		// tpl file
		if ti, ok := rc.Route.Options[KEY_TPL]; ok && ti.(string) != "" && utils.FileExists(ti.(string)) { //定义了tpl文件, 并且文件存在
			tplFile = ti.(string)
		} else if dt := filepath.Join(env.TplDir, rc.Request.URL.Path+".html"); utils.FileExists(dt) { //默认tpl文件, 为: tpldir+url.Path+".html"
			tplFile = dt
		}
		if tplFile != "" {
			if t, err := template.ParseFiles(tplFile); err == nil {
				return t.Execute(rc.Response, data)
			}
		}
	}

	// 以下仍然返回json
	rc.SetHeader("Content-Type", "application/json; charset=UTF-8")
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

/* {{{ func (rc *RESTContext) HTTPOK(data []byte) (err error)
 * 属于request的错误
 */
func (rc *RESTContext) HTTPOK(data []byte) (err error) {
	rc.SetHeader("Content-Type", "text/html; charset=UTF-8")
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
	rc.SetHeader("Content-Type", "image/gif")
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
	rc.SetHeader("Content-Type", "text/html; charset=UTF-8")
	rc.SetHeader("Cache-Control", "max-age=0")
	rc.SetHeader("Cache-Control", "no-cache")
	rc.SetHeader("Cache-Control", "must-revalidate")
	rc.SetHeader("Cache-Control", "private")
	rc.SetHeader("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")
	rc.SetHeader("Pragma", "no-cache")

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
	rc.SetHeader("Content-Type", "text/html; charset=UTF-8")
	rc.SetHeader("Cache-Control", "max-age=0")
	rc.SetHeader("Cache-Control", "no-cache")
	rc.SetHeader("Cache-Control", "must-revalidate")
	rc.SetHeader("Cache-Control", "private")
	rc.SetHeader("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")
	rc.SetHeader("Pragma", "no-cache")
	rc.SetHeader("Location", url)

	_, err = rc.WriteBytes([]byte{})
	return
}

/* }}} */

/* {{{ func (rc *RESTContext) HTTPError(status int) (err error)
 *
 */
func (rc *RESTContext) HTTPError(status int) (err error) {

	rc.SetStatus(status)

	// write data
	err = rc.Output(rc.NewRESTError(status, nil))

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
					rc.SetHeader("Content-Encoding", "gzip")
					b := new(bytes.Buffer)
					w, _ := gzip.NewWriterLevel(b, gzip.BestSpeed)
					w.Write(data)
					w.Close()
					data = b.Bytes()
					break
				} else if val == "deflate" {
					rc.SetHeader("Content-Encoding", "deflate")
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
		rc.SetHeader("Content-Length", strconv.Itoa(rc.ContentLength))
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
	rc.SetHeader("Content-Type", mimetype)
	rc.WriteBytes(data)
}

/* }}} */
