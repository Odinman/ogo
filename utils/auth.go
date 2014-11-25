package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

type Auth struct {
	SecretKeyID    string
	SecretKey      string
	Signature      string
	StringToSign   string
	ClientUniqueID string
	Path           string
	Date           string
	ExpectedSign   string
}

const AuthTag = "dm"

/* {{{ CheckAuth
 */
func (a *Auth) CheckAuth(r *http.Request) error {
	//return nil
	a.ClientUniqueID = r.Header.Get("X-" + AuthTag + "-DeviceId") //不区分大小写
	a.Date = r.Header.Get("Date")
	a.Path = r.URL.Path
	//if a.ClientUniqueID == "" || a.Date == "" {
	//    return errors.New("not enough info")
	//}
	// todo: check date time

	authHeader := r.Header.Get("Authorization")

	//auth的形式为 "{authtag} secretkeyid:signature"
	as := strings.SplitN(authHeader, " ", 2)

	if len(as) != 2 || strings.ToLower(as[0]) != AuthTag {
		return errors.New("require authorization")
	}

	s := strings.SplitN(as[1], ":", 2)

	if len(s) != 2 {
		return errors.New("signature error")
	}

	a.SecretKeyID = s[0]
	signature, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return errors.New("base64 decode error")
	}
	a.Signature = s[1]

	// StringToSign: HTTP-Verb + "\n" + Date + "\n" + HTTP-Path + "\n" + {ClientUniqueID};
	a.StringToSign = r.Method + "\n" + a.Date + "\n" + a.Path + "\n" + a.ClientUniqueID

	a.SecretKey = a.getSecretKey()
	if a.SecretKey == "" {
		return errors.New("get key error")
	}

	mac := hmac.New(sha1.New, []byte(a.SecretKey))
	mac.Write([]byte(a.StringToSign))
	expectedSign := mac.Sum(nil)
	a.ExpectedSign = base64.StdEncoding.EncodeToString(expectedSign)

	if ok := hmac.Equal(signature, expectedSign); !ok {
		return errors.New("signature error")
	}

	return nil
}

/* }}} */

/* {{{ get secretkey
 */
func (a *Auth) getSecretKey() (sk string) {
	//暂时用算法解决，之后需要完全随机数,从高速缓存中查询
	if a.SecretKeyID == "" {
		return
	}
	sk = LengthenUUID(a.SecretKeyID)
	return
}

/* }}} */
