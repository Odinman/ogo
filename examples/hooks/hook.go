// Package hooks provides ...
package hooks

import (
	"github.com/Odinman/ogo"
)

func init() {
	ogo.PreHook(Authentication())
}
