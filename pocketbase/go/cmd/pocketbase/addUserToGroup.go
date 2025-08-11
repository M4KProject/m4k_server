package main

import (
	"github.com/pocketbase/pocketbase/core"
)

// addUserToGroup handles adding a user to a group via the members collection
func addUserToGroup(e *core.RequestEvent) error {
	app := e.App
	userId := e.Request.PathValue("userId")
	groupId := e.Request.PathValue("groupId")

	log := app.Logger()
	log.Info("Attempting to add user %s to group %s", userId, groupId)

	// Check permissions (Admin role required - role >= 30)
	if err := checkPermission(e, groupId, 30); err != nil {
		return err
	}

	if userId == "" || groupId == "" {
		return e.JSON(400, errorJSON("Missing parameters userId=%s groupId=%s", userId, groupId))
	}

	if _, err := app.FindFirstRecordByData("users", "id", userId); err != nil {
		return e.JSON(404, errorJSON("User not found: %s", userId))
	}

	if _, err := app.FindFirstRecordByData("groups", "id", groupId); err != nil {
		return e.JSON(404, errorJSON("Group not found: %s", groupId))
	}

	// Check if user is already a member of the group
	existingMember, err := app.FindFirstRecordByFilter("members", "user = {:user} && group = {:group}", map[string]any{
		"user":  userId,
		"group": groupId,
	})
	if err == nil && existingMember != nil {
		return e.JSON(200, map[string]any{
			"user":  userId,
			"group": groupId,
		})
	}

	memberCollection, err := app.FindCollectionByNameOrId("members")
	if err != nil {
		log.Error("[ADD_USER] Failed to find members collection: %v", err.Error())
		return e.JSON(500, errorJSON("Failed to find members collection"))
	}

	memberRecord := core.NewRecord(memberCollection)
	memberRecord.Set("user", userId)
	memberRecord.Set("group", groupId)

	if err := app.Save(memberRecord); err != nil {
		log.Error("Failed to add user to group: %v", err.Error())
		return e.JSON(500, errorJSON("Failed to add user to group: %v", err.Error()))
	}

	log.Info("Successfully added user %s to group %s", userId, groupId)
	return e.JSON(200, map[string]any{
		"user":  userId,
		"group": groupId,
	})
}
