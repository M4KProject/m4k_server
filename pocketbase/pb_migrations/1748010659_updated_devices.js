/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_2153001328")

  // remove field
  collection.fields.removeById("json3626513111")

  // remove field
  collection.fields.removeById("json325763347")

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_2153001328")

  // add field
  collection.fields.addAt(15, new Field({
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
  collection.fields.addAt(16, new Field({
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
})
