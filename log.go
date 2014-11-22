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
	Debugger.WirteRightNow(msg, logs.LevelCritical)
}

func generateFmtStr(n int) string {
	return strings.Repeat("%v ", n)
}
