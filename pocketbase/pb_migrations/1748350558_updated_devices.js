/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_2153001328")

  // remove field
  collection.fields.removeById("relation1841317061")

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_2153001328")

  // add field
  collection.fields.addAt(20, new Field({
    "cascadeDelete": false,
    "collectionId": "sika7xbbfnwnamj",
    "hidden": false,
    "id": "relation1841317061",
    "maxSelect": 1,
    "minSelect": 0,
    "name": "group",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "relation"
  }))

  return app.save(collection)
})
