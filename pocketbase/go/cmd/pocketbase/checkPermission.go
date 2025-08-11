package main

import (
	"github.com/pocketbase/pocketbase/core"
)

// checkGroupPermission verifies that the authenticated user has the required role in the specified group
func checkPermission(e *core.RequestEvent, groupId string, minRole int) error {
	app := e.App
	auth := e.Auth
	log := app.Logger()

	if auth == nil {
		log.Warn("authentication required")
		return e.JSON(401, errorJSON("authentication required"))
	}

	// Check if user is member of the group with the required role
	member, err := app.FindFirstRecordByFilter("members", "user = {:user} && group = {:group}", map[string]any{
		"user":  auth.Id,
		"group": groupId,
	})
	if err != nil {
		log.Warn("user %s is not a member of group %s", auth.Id, groupId)
		return e.JSON(403, errorJSON("user is not a member of group %s", groupId))
	}

	userRole := member.GetInt("role")
	if userRole < minRole {
		log.Warn("insufficient permissions. user %s group %s role %d", auth.Id, groupId, minRole)
		return e.JSON(403, errorJSON("insufficient permissions"))
	}

	return nil
}
