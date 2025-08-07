// serve.go
package main

import (
	"log"
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
				return e.JSON(500, map[string]string{"error": "Failed to load collections"})
			}
			return e.JSON(200, collections)
		})

		e.Router.GET("/api/pair/{key}/{groupId}", func(e *core.RequestEvent) error {
			key := e.Request.PathValue("key")
			groupId := e.Request.PathValue("groupId")

			log.Printf("[PAIR] Attempting to pair device key=%s with group=%s", key, groupId)

			if key == "" || groupId == "" {
				log.Printf("[PAIR] Error: Missing parameters key=%s groupId=%s", key, groupId)
				return e.JSON(400, map[string]string{"error": "Key and groupId are required"})
			}

			// Find device by key
			log.Printf("[PAIR] Searching for device with key: %s", key)
			device, err := app.FindFirstRecordByFilter("devices", "key = {:key}", map[string]any{
				"key": key,
			})
			if err != nil {
				log.Printf("[PAIR] Device not found for key: %s, error: %v", key, err)
				return e.JSON(404, map[string]string{"error": "Device not found"})
			}

			log.Printf("[PAIR] Found device: %s, current group: %s", device.Id, device.GetString("group"))

			// Check if device already has a group
			currentGroup := device.GetString("group")
			if currentGroup != "" {
				log.Printf("[PAIR] Device already paired to group: %s", currentGroup)
				return e.JSON(400, map[string]string{"error": "Device already paired", "currentGroup": currentGroup})
			}

			// Update device group
			log.Printf("[PAIR] Setting device group to: %s", groupId)
			device.Set("group", groupId)
			if err := app.Save(device); err != nil {
				log.Printf("[PAIR] Failed to save device: %v", err)
				return e.JSON(500, map[string]string{"error": "Failed to pair device"})
			}

			log.Printf("[PAIR] Successfully paired device %s to group %s", key, groupId)
			return e.JSON(200, map[string]string{
				"message": "Device paired successfully",
				"key":     key,
				"group":   groupId,
			})
		})

		// serves static files from the provided public dir (if exists)
		e.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		return e.Next()
	})
}
