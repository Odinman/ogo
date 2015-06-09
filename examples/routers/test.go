// zhaoonline storage service
package routers

import (
	"github.com/Odinman/ogo"

	"../models"
)

type TestRouter struct {
	ogo.Router
}

func init() {
	r := ogo.NewRouter(new(TestRouter), "test").(*TestRouter)
	r.GenericRoute(new(models.Test), ogo.GA_ALL)
	r.Init()
}
