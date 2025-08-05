/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("sika7xbbfnwnamj")

  // remove field
  collection.fields.removeById("text196455508")

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("sika7xbbfnwnamj")

  // add field
  collection.fields.addAt(5, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text196455508",
    "max": 0,
    "min": 0,
    "name": "desc",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  return app.save(collection)
})
