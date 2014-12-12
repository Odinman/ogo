package logs

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
)

const (
	// log message levels
	LevelTrace = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
)

type loggerType func() LoggerInterface

// LoggerInterface defines the behavior of a log provider.
type LoggerInterface interface {
	Init(config string) error
	WriteMsg(msg string, level int) error
	Destroy()
	Flush()
}

var adapters = make(map[string]loggerType)

// Register makes a log provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, log loggerType) {
	if log == nil {
		panic("logs: Register provide is nil")
	}
	if _, dup := adapters[name]; dup {
		panic("logs: Register called twice for provider " + name)
	}
	adapters[name] = log
}

type OLogger struct {
	lock                sync.Mutex
	level               int
	prefix              string
	enableFuncCallDepth bool
	loggerFuncCallDepth int
	msg                 chan *logMsg
	outputs             map[string]LoggerInterface
}

type logMsg struct {
	level int
	msg   string
}

// NewLogger returns a new OLogger.
// channellen means the number of messages in chan.
// if the buffering chan is full, logger adapters write to file or other way.
func NewLogger(channellen int64) *OLogger {
	bl := new(OLogger)
	bl.loggerFuncCallDepth = 2
	bl.msg = make(chan *logMsg, channellen)
	bl.outputs = make(map[string]LoggerInterface)
	//bl.SetLogger("console", "") // default output to console
	go bl.startLogger()
	return bl
}

// SetLogger provides a given logger adapter into OLogger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (bl *OLogger) SetLogger(adaptername string, config string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if log, ok := adapters[adaptername]; ok {
		lg := log()
		err := lg.Init(config)
		bl.outputs[adaptername] = lg
		if err != nil {
			fmt.Println("logs.OLogger.SetLogger: " + err.Error())
			return err
		}
	} else {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adaptername)
	}
	return nil
}

// remove a logger adapter in OLogger.
func (bl *OLogger) DelLogger(adaptername string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if lg, ok := bl.outputs[adaptername]; ok {
		lg.Destroy()
		delete(bl.outputs, adaptername)
		return nil
	} else {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adaptername)
	}
}

func (bl *OLogger) writerMsg(loglevel int, msg string) error {
	if bl.level > loglevel {
		return nil
	}
	lm := new(logMsg)
	lm.level = loglevel
	if bl.enableFuncCallDepth {
		_, file, line, ok := runtime.Caller(bl.loggerFuncCallDepth)
		if ok {
			_, filename := path.Split(file)
			lm.msg = fmt.Sprintf("%s [%s:%d]", msg, filename, line)
		} else {
			lm.msg = msg
		}
	} else {
		lm.msg = msg
	}
	bl.msg <- lm
	return nil
}

// set log message level.
// if message level (such as LevelTrace) is less than logger level (such as LevelWarn), ignore message.
func (bl *OLogger) SetLevel(l int) {
	bl.level = l
}

// set log prefix
func (bl *OLogger) SetPrefix(p string) {
	bl.prefix = p
}

// set log funcCallDepth
func (bl *OLogger) SetLogFuncCallDepth(d int) {
	bl.loggerFuncCallDepth = d
}

// enable log funcCallDepth
func (bl *OLogger) EnableFuncCallDepth(b bool) {
	bl.enableFuncCallDepth = b
}

// start logger chan reading.
// when chan is full, write logs.
func (bl *OLogger) startLogger() {
	for {
		select {
		case bm := <-bl.msg:
			for _, l := range bl.outputs {
				l.WriteMsg(bm.msg, bm.level)
			}
		}
	}
}

// log trace level message.
func (bl *OLogger) Trace(format string, v ...interface{}) {
	if bl.prefix != "" {
		format = bl.prefix + " " + format
	}
	msg := fmt.Sprintf("[T] "+format, v...)
	bl.writerMsg(LevelTrace, msg)
}

// log debug level message.
func (bl *OLogger) Debug(format string, v ...interface{}) {
	if bl.prefix != "" {
		format = bl.prefix + " " + format
	}
	msg := fmt.Sprintf("[D] "+format, v...)
	bl.writerMsg(LevelDebug, msg)
}

// log info level message.
func (bl *OLogger) Info(format string, v ...interface{}) {
	if bl.prefix != "" {
		format = bl.prefix + " " + format
	}
	msg := fmt.Sprintf("[I] "+format, v...)
	bl.writerMsg(LevelInfo, msg)
}

// log warn level message.
func (bl *OLogger) Warn(format string, v ...interface{}) {
	if bl.prefix != "" {
		format = bl.prefix + " " + format
	}
	msg := fmt.Sprintf("[W] "+format, v...)
	bl.writerMsg(LevelWarn, msg)
}

// log error level message.
func (bl *OLogger) Error(format string, v ...interface{}) {
	if bl.prefix != "" {
		format = bl.prefix + " " + format
	}
	msg := fmt.Sprintf("[E] "+format, v...)
	bl.writerMsg(LevelError, msg)
}

// log critical level message.
func (bl *OLogger) Critical(format string, v ...interface{}) {
	if bl.prefix != "" {
		format = bl.prefix + " " + format
	}
	msg := fmt.Sprintf("[C] "+format, v...)
	bl.writerMsg(LevelCritical, msg)
}

// for gorp
func (bl *OLogger) Printf(format string, v ...interface{}) {
	if bl.prefix != "" {
		format = bl.prefix + " " + format
	}
	msg := fmt.Sprintf("[P] "+format, v...)
	bl.writerMsg(LevelDebug, msg) //默认为debug级别
}

func (ol *OLogger) Log(tag, format string, v ...interface{}) {
	switch strings.ToLower(tag) {
	case "trace":
		ol.Trace(format, v...)
	case "debug":
		ol.Debug(format, v...)
	case "info":
		ol.Info(format, v...)
	case "warn":
		ol.Warn(format, v...)
	case "error":
		ol.Error(format, v...)
	case "critical":
		ol.Critical(format, v...)
	}
}

// outputs directly.
func (bl *OLogger) WirteRightNow(msg string, level int) {
	for _, l := range bl.outputs {
		l.WriteMsg(msg, level)
	}
}

// flush all chan data.
func (bl *OLogger) Flush() {
	for _, l := range bl.outputs {
		l.Flush()
	}
}

// close logger, flush all chan data and destroy all adapters in OLogger.
func (bl *OLogger) Close() {
	for {
		if len(bl.msg) > 0 {
			bm := <-bl.msg
			for _, l := range bl.outputs {
				l.WriteMsg(bm.msg, bm.level)
			}
		} else {
			break
		}
	}
	for _, l := range bl.outputs {
		l.Flush()
		l.Destroy()
	}
}
