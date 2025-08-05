/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_3572739349")

  // update collection data
  unmarshal({
    "indexes": [
      "CREATE INDEX `idx_IaXV31pQAo` ON `members` (\n  `owner`,\n  `group`\n)"
    ]
  }, collection)

  // update field
  collection.fields.addAt(6, new Field({
    "cascadeDelete": false,
    "collectionId": "_pb_users_auth_",
    "hidden": false,
    "id": "relation2375276105",
    "maxSelect": 1,
    "minSelect": 0,
    "name": "owner",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "relation"
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_3572739349")

  // update collection data
  unmarshal({
    "indexes": [
      "CREATE UNIQUE INDEX `members_user_group_unique` ON `members` (\n  `user`,\n  `group`\n)"
    ]
  }, collection)

  // update field
  collection.fields.addAt(6, new Field({
    "cascadeDelete": false,
    "collectionId": "_pb_users_auth_",
    "hidden": false,
    "id": "relation2375276105",
    "maxSelect": 1,
    "minSelect": 0,
    "name": "user",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "relation"
  }))

  return app.save(collection)
})
