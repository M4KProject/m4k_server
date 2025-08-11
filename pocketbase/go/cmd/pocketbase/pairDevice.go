package main

import (
	"github.com/pocketbase/pocketbase/core"
)

func pairDevice(e *core.RequestEvent) error {
	app := e.App
	key := e.Request.PathValue("key")
	groupId := e.Request.PathValue("groupId")

	log := app.Logger()

	if err := checkPermission(e, groupId, 30); err != nil {
		return err
	}

	log.Info("Attempting to pair device key=%s with group=%s", key, groupId)

	if key == "" || groupId == "" {
		log.Warn("Missing parameters key=%s groupId=%s", key, groupId)
		return e.JSON(400, errorJSON("Missing parameters"))
	}

	device, err := app.FindFirstRecordByFilter("devices", "key = {:key}", map[string]any{
		"key": key,
	})
	if err != nil {
		log.Warn("Device not found for key: %s, error: %v", key, err.Error())
		return e.JSON(404, errorJSON("Device not found"))
	}

	// Check if device already has a group
	currentGroup := device.GetString("group")
	if currentGroup != "" {
		log.Warn("Device already paired")
		return e.JSON(400, errorJSON("Device already paired"))
	}

	// Verify group exists
	_, err = app.FindFirstRecordByData("groups", "id", groupId)
	if err != nil {
		log.Warn("Group not found: %s, error: %v", groupId, err.Error())
		return e.JSON(404, errorJSON("Group not found"))
	}

	// Update device group
	device.Set("group", groupId)

	if err := app.Save(device); err != nil {
		log.Error("Failed to pair device, error: %s", err.Error())
		return e.JSON(500, errorJSON("Failed to pair device"))
	}

	log.Info("Successfully paired device %s to group %s", key, groupId)
	return e.JSON(200, map[string]string{
		"key":   key,
		"group": groupId,
	})
}
