// Package hooks provides ...
package hooks

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/Odinman/ogo"
	"github.com/Odinman/ogo/utils"

	. "../config"
)

type Auth struct {
	SecretKeyID    string //安全密钥id
	SecretKey      string //安全密钥
	Signature      string //签名
	StringToSign   string // 待签名字符
	ClientUniqueID string // 客户端独立标识
	Path           string // url path
	Date           string // request date
	ExpectedSign   string // 正确签名
}

var xDmUserId = http.CanonicalHeaderKey("X-QH-UserId")

func init() {
	ogo.PreHook(Authentication())
}

/* {{{  func Authentication() ogo.OgoHook
 *
 */
func Authentication() ogo.OgoHook {

	fn := func(c *ogo.RESTContext) error {
		// check if login
		c.Debug("route pattern: %s", c.Route.Pattern)

		// 从header里找userid
		userId := c.Request.Header.Get(xDmUserId)
		if userId != "" { // userid不为空, 可能已经登录
			c.SetEnv("_userid", userId)
		} else if SkipLogin { //没登录
			// 全局设置可以不用登录
			c.Info("app setting to skiplogin")
		} else if sl, ok := c.Route.Options[ogo.KEY_SKIPLOGIN]; !ok {
			//没有设置skip,需要登录
			c.Warn("Need login, method: %s, pattern: %s", c.Route.Method, c.Route.Pattern)
			return c.NewRESTError(http.StatusUnauthorized, "Please Login")
		} else if sl == true {
			//设置了skip
			c.Info("Not need login, method: %s, pattern: %s", c.Route.Method, c.Route.Pattern)
		} else {
			c.Warn("Need login, method: %s, pattern: %s", c.Route.Method, c.Route.Pattern)
			return c.NewRESTError(http.StatusUnauthorized, "Please Login")
		}

		// authentication
		if SkipAuth {
			c.Info("app setting to skipauth")
		} else {
			auth := new(Auth)
			if err := auth.check(c); err != nil {
				c.Info("ogo auth failed, id: %s, key: %s, signature: %s, expected: %s", auth.SecretKeyID, auth.SecretKey, auth.Signature, auth.ExpectedSign)
				c.Error("auth error: %s", err)
				if sa, ok := c.Route.Options[ogo.KEY_SKIPAUTH]; ok {
					if sa == true {
						c.Info("Route(%s %s) not need auth", c.Route.Method, c.Route.Pattern)
						return nil
					}
				}
				return c.NewRESTError(http.StatusUnauthorized, err)
			}
		}

		return nil
	}

	return ogo.OgoHook(fn)
}

/* }}} */

func (a *Auth) check(c *ogo.RESTContext) error {
	r := c.Request
	a.ClientUniqueID = r.Header.Get(xDmUserId)
	a.Date = r.Header.Get("Date")
	a.Path = r.URL.Path

	authHeader := r.Header.Get("Authorization")

	//auth的形式为 "DM secretkeyid:signature"
	as := strings.SplitN(authHeader, " ", 2)

	if len(as) != 2 || strings.ToLower(as[0]) != "dm" {
		return c.NewRESTError(http.StatusUnauthorized, "require authorization header")
	}

	s := strings.SplitN(as[1], ":", 2)

	if len(s) != 2 {
		return c.NewRESTError(http.StatusUnauthorized, "signature error")
	}

	a.SecretKeyID = s[0]
	signature, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return c.NewRESTError(http.StatusUnauthorized, "base64 decode error")
	}
	a.Signature = s[1]

	// StringToSign: HTTP-Verb + "\n" + Date + "\n" + HTTP-Path + "\n" + {ClientUniqueID};
	a.StringToSign = r.Method + "\n" + a.Date + "\n" + a.Path + "\n" + a.ClientUniqueID

	a.SecretKey = a.getSecretKey()
	if a.SecretKey == "" {
		return c.NewRESTError(http.StatusUnauthorized, "get key error")
	}

	mac := hmac.New(sha1.New, []byte(a.SecretKey))
	mac.Write([]byte(a.StringToSign))
	expectedSign := mac.Sum(nil)
	a.ExpectedSign = base64.StdEncoding.EncodeToString(expectedSign)

	if ok := hmac.Equal(signature, expectedSign); !ok {
		return c.NewRESTError(http.StatusUnauthorized, "signature error")
	}
	return nil
}

/* {{{ get secretkey
 */
func (a *Auth) getSecretKey() (sk string) {
	//暂时用算法解决，之后需要完全随机数,从高速缓存中查询
	if a.SecretKeyID == "" {
		return
	}
	sk = utils.LengthenUUID(a.SecretKeyID)
	return
}

/* }}} */
