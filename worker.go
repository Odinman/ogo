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
