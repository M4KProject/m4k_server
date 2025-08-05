/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_3446931122")

  // update collection data
  unmarshal({
    "name": "assets"
  }, collection)

  // remove field
  collection.fields.removeById("json3414765911")

  // add field
  collection.fields.addAt(8, new Field({
    "hidden": false,
    "id": "number2350531887",
    "max": null,
    "min": null,
    "name": "width",
    "onlyInt": false,
    "presentable": false,
    "required": false,
    "system": false,
    "type": "number"
  }))

  // add field
  collection.fields.addAt(9, new Field({
    "hidden": false,
    "id": "number4115522831",
    "max": null,
    "min": null,
    "name": "height",
    "onlyInt": false,
    "presentable": false,
    "required": false,
    "system": false,
    "type": "number"
  }))

  // update field
  collection.fields.addAt(6, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text1579384326",
    "max": 0,
    "min": 0,
    "name": "path",
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
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_3446931122")

  // update collection data
  unmarshal({
    "name": "files"
  }, collection)

  // add field
  collection.fields.addAt(8, new Field({
    "hidden": false,
    "id": "json3414765911",
    "maxSize": 0,
    "name": "info",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "json"
  }))

  // remove field
  collection.fields.removeById("number2350531887")

  // remove field
  collection.fields.removeById("number4115522831")

  // update field
  collection.fields.addAt(6, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text1579384326",
    "max": 0,
    "min": 0,
    "name": "name",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  // update field
  collection.fields.addAt(9, new Field({
    "hidden": false,
    "id": "file1542800728",
    "maxSelect": 99,
    "maxSize": 0,
    "mimeTypes": [],
    "name": "field",
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
