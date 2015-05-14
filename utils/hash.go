// Package utils provides ...
package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

/* {{{ func HashSha1(oirg, salt string) string
 * 密码加密
 */
func HashSha1(oirg, salt string) string {
	mac := hmac.New(sha1.New, []byte(salt))
	mac.Write([]byte(oirg))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

/* }}} */
