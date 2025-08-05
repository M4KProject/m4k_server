/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_4105455982")

  // update field
  collection.fields.addAt(4, new Field({
    "autogeneratePattern": "[a-z][a-z0-9]{5}",
    "hidden": false,
    "id": "text2324736937",
    "max": 256,
    "min": 3,
    "name": "key",
    "pattern": "^[a-z][a-z0-9]+$",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_4105455982")

  // update field
  collection.fields.addAt(4, new Field({
    "autogeneratePattern": "[a-z][a-z0-9]{6}",
    "hidden": false,
    "id": "text2324736937",
    "max": 256,
    "min": 3,
    "name": "key",
    "pattern": "^[a-z][a-z0-9]+$",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  return app.save(collection)
})
