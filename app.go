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
 * 设置环境变量
 */
func (rc *RESTContext) NewAppLogging(al *AppLog) {
	rc.Access.SaveApp(al)
}

/* }}} */

/* {{{ func (rc *RESTContext) AppLoggingNew(i interface{})
 * 设置环境变量
 */
func (rc *RESTContext) AppLoggingNew(i interface{}) {
	rc.Access.App.New = i
}

/* }}} */

/* {{{ func (rc *RESTContext) AppLoggingOld(i interface{})
 * 设置环境变量
 */
func (rc *RESTContext) AppLoggingOld(i interface{}) {
	rc.Access.App.Old = i
}

/* }}} */

/* {{{ func (rc *RESTContext) AppLoggingResult(i interface{})
 * 设置环境变量
 */
func (rc *RESTContext) AppLoggingResult(i interface{}) {
	rc.Access.App.Result = i
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
func (rc *RESTContext) SetOTP(v, t, s string) {
	otp := new(OTPSpec)
	otp.Value = v
	otp.Type = t
	otp.Sn = s
	rc.OTP = otp
}

/* }}} */
