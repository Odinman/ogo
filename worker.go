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
    Ctx        *Context
}

type WorkerInterface interface {
    Init(ctx *Context, name string)
    Main() error
}

// 初始化, 赋值
func (w *Worker) Init(ctx *Context, name string) {
    w.Ctx = ctx
    w.WorkerName = name
    //Debugger.Debug("init")
}

func (w *Worker) Main() error {
    Debugger.Warn("no main function")
    return errors.New("no main function")
}

// worker registor
func AddWorker(w WorkerInterface) {
    reflectVal := reflect.ValueOf(w)
    ct := reflect.Indirect(reflectVal).Type()
    wName := strings.ToLower(strings.TrimSuffix(ct.Name(), "Worker"))
    if _, ok := Ctx.Workers[wName]; ok {
        //worker不能重名
        return
    } else {
        worker := &Worker{}
        worker.WorkerType = ct
        Ctx.Workers[wName] = worker
    }
}
