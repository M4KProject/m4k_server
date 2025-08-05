/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_3446931122")

  // update collection data
  unmarshal({
    "name": "files"
  }, collection)

  // add field
  collection.fields.addAt(10, new Field({
    "hidden": false,
    "id": "file1602912115",
    "maxSelect": 1,
    "maxSize": 0,
    "mimeTypes": [],
    "name": "source",
    "presentable": false,
    "protected": false,
    "required": false,
    "system": false,
    "thumbs": [
      "24x24",
      "48x48",
      "100x100",
      "200x200",
      "360x360",
      "720x720",
      "1920x1920",
      "4096x4096"
    ],
    "type": "file"
  }))

  // update field
  collection.fields.addAt(5, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2363381545",
    "max": 0,
    "min": 0,
    "name": "mimetype",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  // update field
  collection.fields.addAt(11, new Field({
    "hidden": false,
    "id": "file1542800728",
    "maxSelect": 99,
    "maxSize": 0,
    "mimeTypes": [],
    "name": "formats",
    "presentable": false,
    "protected": false,
    "required": false,
    "system": false,
    "thumbs": [
      "24x24",
      "48x48",
      "100x100",
      "200x200",
      "360x360",
      "720x720",
      "1920x1920",
      "4096x4096"
    ],
    "type": "file"
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_3446931122")

  // update collection data
  unmarshal({
    "name": "assets"
  }, collection)

  // remove field
  collection.fields.removeById("file1602912115")

  // update field
  collection.fields.addAt(5, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2363381545",
    "max": 0,
    "min": 0,
    "name": "type",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  // update field
  collection.fields.addAt(10, new Field({
    "hidden": false,
    "id": "file1542800728",
    "maxSelect": 99,
    "maxSize": 0,
    "mimeTypes": [],
    "name": "files",
    "presentable": false,
    "protected": false,
    "required": false,
    "system": false,
    "thumbs": [
      "24x24",
      "48x48",
      "100x100",
      "200x200",
      "360x360",
      "720x720",
      "1920x1920",
      "4096x4096"
    ],
    "type": "file"
  }))

  return app.save(collection)
})
