// serve.go
package main

import (
	"os"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func bindServe(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		e.Router.GET("/api/now", func(e *core.RequestEvent) error {
			return e.String(200, time.Now().Format(time.RFC3339))
		})

		e.Router.GET("/api/schema", func(e *core.RequestEvent) error {
			collections, err := app.FindAllCollections()
			if err != nil {
				return e.JSON(500, errorJSON("Failed to load collections"))
			}
			return e.JSON(200, collections)
		})

		// Device pairing functionality
		e.Router.GET("/api/pair/{key}/{groupId}", pairDevice).Bind(apis.RequireAuth())

		// User management functionality
		e.Router.POST("/api/groups/{groupId}/members/{userId}", addUserToGroup).Bind(apis.RequireAuth())

		// serves static files from the provided public dir (if exists)
		e.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		return e.Next()
	})
}
