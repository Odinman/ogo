// Package utils provides ...
package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

/* {{{ func HashSha1(orig, salt string) string
 * 密码加密
 */
func HashSha1(i ...string) string {
	if len(i) <= 0 {
		return ""
	}
	var orig, salt string
	orig = i[0]
	if len(i) >= 2 {
		salt = i[1]
	} else {
		salt = "odin"
	}
	mac := hmac.New(sha1.New, []byte(salt))
	mac.Write([]byte(orig))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

/* }}} */
