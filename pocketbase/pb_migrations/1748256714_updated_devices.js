/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_2153001328")

  // remove field
  collection.fields.removeById("text3626513111")

  // remove field
  collection.fields.removeById("text325763347")

  // add field
  collection.fields.addAt(16, new Field({
    "hidden": false,
    "id": "json3626513111",
    "maxSize": 0,
    "name": "input",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "json"
  }))

  // add field
  collection.fields.addAt(17, new Field({
    "hidden": false,
    "id": "json325763347",
    "maxSize": 0,
    "name": "result",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "json"
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_2153001328")

  // add field
  collection.fields.addAt(16, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text3626513111",
    "max": 0,
    "min": 0,
    "name": "input",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  // add field
  collection.fields.addAt(17, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text325763347",
    "max": 0,
    "min": 0,
    "name": "result",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  // remove field
  collection.fields.removeById("json3626513111")

  // remove field
  collection.fields.removeById("json325763347")

  return app.save(collection)
})
