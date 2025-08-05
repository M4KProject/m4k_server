/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("sika7xbbfnwnamj")

  // update field
  collection.fields.addAt(3, new Field({
    "autogeneratePattern": "[a-z][a-z0-9]{4}",
    "hidden": false,
    "id": "text2324736937",
    "max": 0,
    "min": 0,
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
  const collection = app.findCollectionByNameOrId("sika7xbbfnwnamj")

  // update field
  collection.fields.addAt(3, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2324736937",
    "max": 0,
    "min": 0,
    "name": "key",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  return app.save(collection)
})
