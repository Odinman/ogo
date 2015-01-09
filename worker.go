// Ogo

package ogo

import (
	"errors"
	"reflect"
	"strings"
)

type Worker struct {
	WorkerName string
	WorkerType reflect.Type
	Mux        *Mux
}

type WorkerInterface interface {
	Init(mux *Mux, name string)
	Main() error
}

// 初始化, 赋值
func (w *Worker) Init(mux *Mux, name string) {
	w.Mux = mux
	w.WorkerName = name
}

func (w *Worker) Main() error {
	Warn("no main function")
	return errors.New("no main function")
}

// worker registor
func AddWorker(w WorkerInterface) {
	reflectVal := reflect.ValueOf(w)
	ct := reflect.Indirect(reflectVal).Type()
	wName := strings.ToLower(strings.TrimSuffix(ct.Name(), "Worker"))
	if _, ok := DMux.Workers[wName]; ok {
		//worker不能重名
		return
	} else {
		worker := &Worker{}
		worker.WorkerType = ct
		DMux.Workers[wName] = worker
	}
}

/* {{{	Worker loggers
 * 可以在每个debug信息带上session
 */
func (w *Worker) Trace(format string, v ...interface{}) {
	w.logf("trace", format, v...)
}
func (w *Worker) Debug(format string, v ...interface{}) {
	w.logf("debug", format, v...)
}
func (w *Worker) Info(format string, v ...interface{}) {
	w.logf("info", format, v...)
}
func (w *Worker) Print(format string, v ...interface{}) {
	w.logf("info", format, v...)
}
func (w *Worker) Warn(format string, v ...interface{}) {
	w.logf("warn", format, v...)
}
func (w *Worker) Error(format string, v ...interface{}) {
	w.logf("error", format, v...)
}
func (w *Worker) Critical(format string, v ...interface{}) {
	w.logf("critical", format, v...)
}
func (w *Worker) logf(tag, format string, v ...interface{}) {
	//var prefix string
	//if p := w.GetEnv(LogPrefixKey); p != nil {
	//	prefix = p.(string)
	//}
	//if prefix != "" {
	//	format = prefix + " " + format
	//}
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
