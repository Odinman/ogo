// Package ogo provides ...
package ogo

import (
	"reflect"
	"sync"
)

type OgoHook func(c *RESTContext) error

type HStack struct {
	lock      sync.Mutex
	preHooks  []OgoHook
	postHooks []OgoHook
}

// struct里面的field可定义处理函数
type TagHook func(v reflect.Value) reflect.Value
