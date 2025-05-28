/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("sika7xbbfnwnamj")

  // update collection data
  unmarshal({
    "createRule": "@request.auth.id != \"\"",
    "deleteRule": "@request.auth.id != \"\"",
    "listRule": "@request.auth.id != \"\"",
    "updateRule": "@request.auth.id != \"\"",
    "viewRule": "@request.auth.id != \"\""
  }, collection)

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("sika7xbbfnwnamj")

  // update collection data
  unmarshal({
    "createRule": "@request.auth.id != \"\" && owner.id = @request.auth.id",
    "deleteRule": "@request.auth.id != \"\" && owner.id = @request.auth.id",
    "listRule": "@request.auth.id != \"\" && owner.id = @request.auth.id",
    "updateRule": "@request.auth.id != \"\" && owner.id = @request.auth.id",
    "viewRule": "@request.auth.id != \"\" && owner.id = @request.auth.id"
  }, collection)

  return app.save(collection)
})
