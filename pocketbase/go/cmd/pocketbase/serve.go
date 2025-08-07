// serve.go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func errorJSON(error string) any {
	return map[string]string{"error": error}
}

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

		e.Router.GET("/api/pair/{key}/{groupId}", func(e *core.RequestEvent) error {
			key := e.Request.PathValue("key")
			groupId := e.Request.PathValue("groupId")

			log := app.Logger()

			msg := fmt.Sprintf("Attempting to pair device key=%s with group=%s", key, groupId)
			log.Info("[PAIR] " + msg)

			if key == "" || groupId == "" {
				msg = fmt.Sprintf("Error: Missing parameters key=%s groupId=%s", key, groupId)
				log.Warn("[PAIR] " + msg)
				return e.JSON(400, errorJSON(msg))
			}

			device, err := app.FindFirstRecordByFilter("devices", "key = {:key}", map[string]any{
				"key": key,
			})
			if err != nil {
				msg = fmt.Sprintf("Device not found for key: %s, error: %v", key, err.Error())
				log.Warn("[PAIR] " + msg)
				return e.JSON(404, errorJSON(msg))
			}

			// Check if device already has a group
			currentGroup := device.GetString("group")
			if currentGroup != "" {
				msg = fmt.Sprintf("Device already paired key: %s", key)
				log.Warn("[PAIR] " + msg)
				return e.JSON(400, errorJSON(msg))
			}

			// Verify group exists
			_, err = app.FindFirstRecordByData("groups", "id", groupId)
			if err != nil {
				msg = fmt.Sprintf("Group not found: %s, error: %v", groupId, err.Error())
				log.Warn("[PAIR] " + msg)
				return e.JSON(404, errorJSON(msg))
			}

			// Update device group
			device.Set("group", groupId)

			if err := app.Save(device); err != nil {
				msg = fmt.Sprintf("Failed to pair device, error: %v", err.Error())
				log.Error("[PAIR] " + msg)
				return e.JSON(500, errorJSON(msg))
			}

			msg = fmt.Sprintf("Successfully paired device %s to group %s", key, groupId)
			log.Info("[PAIR] " + msg)
			return e.JSON(200, map[string]string{
				"key":   key,
				"group": groupId,
			})
		})

		// serves static files from the provided public dir (if exists)
		e.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		return e.Next()
	})
}
