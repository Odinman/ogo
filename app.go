package ogo

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

/* {{{ func (rc *RESTContext) NewAppLogging(al *AppLog)
 * 新建App日志
 */
func (rc *RESTContext) NewAppLogging(al *AppLog) {
	rc.Access.SaveApp(al)
}

/* }}} */

/* {{{ func (rc *RESTContext) AppLoggingNew(i interface{})
 * 设置环境变量
 */
func (rc *RESTContext) AppLoggingNew(i interface{}) {
	rc.Access.App.(*AppLog).New = i
}

/* }}} */

/* {{{ func (rc *RESTContext) AppLoggingCtag(ctag string)
 * 新
 */
func (rc *RESTContext) AppLoggingCtag(ctag string) {
	rc.Access.App.(*AppLog).Ctag = ctag
}

/* }}} */

/* {{{ func (rc *RESTContext) AppLoggingOld(i interface{})
 * 旧
 */
func (rc *RESTContext) AppLoggingOld(i interface{}) {
	rc.Access.App.(*AppLog).Old = i
}

/* }}} */

/* {{{ func (rc *RESTContext) AppLoggingResult(i interface{})
 * 结果日志
 */
func (rc *RESTContext) AppLoggingResult(i interface{}) {
	rc.Access.App.(*AppLog).Result = i
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

/* {{{ func (rc *RESTContext) SetOTP(v,t,s string)
 * 设置环境变量
 */
func (rc *RESTContext) SetOTP(s ...string) {
	if len(s) > 0 {
		otp := new(OTPSpec)
		otp.Value = s[0]
		if len(s) > 1 {
			otp.Type = s[1]
		}
		if len(s) > 2 {
			otp.Sn = s[2]
		}
		rc.OTP = otp
	}
}

/* }}} */
