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
	debugger.WirteRightNow(msg, logs.LevelCritical)
}

func generateFmtStr(n int) string {
	return strings.Repeat("%v ", n)
}

func Trace(format string, v ...interface{}) {
	debugger.Trace(format, v...)
}
func Debug(format string, v ...interface{}) {
	debugger.Debug(format, v...)
}
func Info(format string, v ...interface{}) {
	debugger.Info(format, v...)
}
func Warn(format string, v ...interface{}) {
	debugger.Warn(format, v...)
}
func Error(format string, v ...interface{}) {
	debugger.Error(format, v...)
}
func Critical(format string, v ...interface{}) {
	debugger.Critical(format, v...)
}
func Log(level, format string, v ...interface{}) {
	debugger.Log(level, format, v...)
}
