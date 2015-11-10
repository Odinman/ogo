package ogo

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"net/http"
	"strconv"
	"strings"

	//"github.com/zenazn/goji/web"
)

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
