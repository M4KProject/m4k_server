package main

import (
	"log"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/spf13/cobra"
)

func toRule(rule string) *string {
	log.Println("  toRule", rule)
	rule = " " + rule + " "
	rule = strings.ReplaceAll(rule, " auth ", " @request.auth.id ")
	rule = strings.ReplaceAll(rule, " group_members.", " group.members_via_group.")
	rule = strings.ReplaceAll(rule, " members.", " members_via_group.")
	rule = strings.ReplaceAll(rule, "\t", " ")
	rule = strings.ReplaceAll(rule, "\n", " ")
	rule = strings.ReplaceAll(rule, "  ", " ")
	rule = strings.ReplaceAll(rule, "  ", " ")
	rule = strings.ReplaceAll(rule, "( ", "(")
	rule = strings.ReplaceAll(rule, " )", ")")
	rule = strings.Trim(rule, " ")
	log.Println("  toRule result", rule)
	return types.Pointer(rule)
}

func initRule(app *pocketbase.PocketBase, coll *core.Collection) {
	log.Println("Collection Rule : " + coll.Name)

	if strings.HasPrefix(coll.Name, "_") {
		return
	}

	if coll.Name == "users" {
		return
	}

	viewRule := ""
	editRule := ""

	viewRule = `auth != "" && group_members.user ?= auth && group_members.role ?>= 10`
	editRule = `auth != "" && group_members.user ?= auth && group_members.role ?>= 20`

	if coll.Name == "devices" {
		viewRule = `auth != "" && (user = auth || group_members.user ?= auth && group_members.role ?>= 10)`
		editRule = `auth != "" && (user = auth || group_members.user ?= auth && group_members.role ?>= 20)`
	}

	if coll.Name == "members" {
		viewRule = `auth != "" && (group.user = auth || group_members.user ?= auth && group_members.role ?>= 10)`
		editRule = `auth != "" && (group.user = auth || group_members.user ?= auth && group_members.role ?>= 30)`
	}

	if coll.Name == "groups" {
		viewRule = `auth != "" && (user = auth || members.user ?= auth && members.role ?>= 10)`
		editRule = `auth != "" && (user = auth || members.user ?= auth && members.role ?>= 30)`
	}

	log.Println("  view:" + viewRule)
	log.Println("  edit:" + editRule)

	coll.ListRule = toRule(viewRule)
	coll.ViewRule = toRule(viewRule)
	coll.CreateRule = toRule(editRule)
	coll.UpdateRule = toRule(editRule)
	coll.DeleteRule = toRule(editRule)

	if coll.Name == "groups" {
		coll.CreateRule = toRule(`auth != "" && user = auth`)
	}

	err := app.Save(coll)
	if err != nil {
		log.Println("  view:" + viewRule)
		log.Println("  edit:" + editRule)
		log.Println("Error:" + err.Error())
	}
}

func initRules(app *pocketbase.PocketBase) {
	collections, err := app.FindAllCollections()

	if err != nil {
		log.Println("Error:" + err.Error())
		return
	}

	for _, coll := range collections {
		initRule(app, coll)
	}
}

func main() {
	app := pocketbase.New()

	app.RootCmd.AddCommand(&cobra.Command{
		Use: "init",
		Run: func(cmd *cobra.Command, args []string) {
			initRules(app)
		},
	})

	if err := app.Start(); err != nil {
		panic(err)
	}
}
