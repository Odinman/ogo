// Package ogo provides ...
package ogo

import ()

type OgoHook func(c *RESTContext)

type HStack struct {
	preHooks  []OgoHook
	postHooks []OgoHook
}
