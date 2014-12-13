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
