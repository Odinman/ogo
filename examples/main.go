// zhaoonline storage service
package main

import (
	"github.com/Odinman/ogo"

	_ "./hooks"
	_ "./routers"
)

func main() {
	ogo.Run()
}
