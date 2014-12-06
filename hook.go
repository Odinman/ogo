// Package ogo provides ...
package ogo

import ()

type OgoHook func(c *RESTContext)

type HStack struct {
	preHooks  []OgoHook
	postHooks []OgoHook
}

var Hooks *HStack

func init() {
	Hooks = &HStack{
		preHooks:  make([]OgoHook, 0),
		postHooks: make([]OgoHook, 0),
	}
}

func PreHook(hook OgoHook) {
	Hooks.preHooks = append(Hooks.preHooks, hook)
}

func PostHook(hook OgoHook) {
	Hooks.preHooks = append(Hooks.postHooks, hook)
}
