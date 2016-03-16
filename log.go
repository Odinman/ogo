// Debugger

package ogo

import (
	"fmt"
	"strings"

	"github.com/Odinman/ogo/libs/logs"
)

//立即输出
func WriteMsg(v ...interface{}) {
	msg := fmt.Sprintf("[F] "+generateFmtStr(len(v)), v...)
	logger.WirteRightNow(msg, logs.LevelCritical)
}

func generateFmtStr(n int) string {
	return strings.Repeat("%v ", n)
}

func Trace(format string, v ...interface{}) {
	logger.Trace(format, v...)
}
func Debug(format string, v ...interface{}) {
	logger.Debug(format, v...)
}
func Info(format string, v ...interface{}) {
	logger.Info(format, v...)
}
func Warn(format string, v ...interface{}) {
	logger.Warn(format, v...)
}
func Error(format string, v ...interface{}) {
	logger.Error(format, v...)
}
func Critical(format string, v ...interface{}) {
	logger.Critical(format, v...)
}
func Log(level, format string, v ...interface{}) {
	logger.Log(level, format, v...)
}

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
