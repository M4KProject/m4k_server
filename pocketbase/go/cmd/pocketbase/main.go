// main.go
package main

import (
	"github.com/pocketbase/pocketbase"
)

func main() {
	app := pocketbase.New()

	bindMedias(app)
	bindServe(app)
	bindJobs(app)

	// Bind du transcodage vidéo
	bindTranscode(app)

	if err := app.Start(); err != nil {
		panic(err)
	}
}
