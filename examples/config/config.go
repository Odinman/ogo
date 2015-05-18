// Package main provides ...
package config

import (
	"github.com/Odinman/ogo"
)

const (
	DB_WRITE = "write"
	DB_READ  = "read"
)

var (
	SkipLogin bool
	SkipAuth  bool
	CDNDomain string
)

func init() {
	//全局忽略登录
	if sl, err := ogo.Config().Bool("SkipLogin"); err == nil {
		SkipLogin = sl
	}
	//全局忽略鉴权
	if sa, err := ogo.Config().Bool("SkipAuth"); err == nil {
		SkipAuth = sa
	}

	//全局忽略鉴权
	if domain := ogo.Config().String("CDNDomain"); domain != "" {
		CDNDomain = domain
	}
}
