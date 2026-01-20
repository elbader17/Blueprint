# My API Blueprint

```json
{
  "project_name": "MyAwesomeAPI",
  "database": {
    "type": "firestore",
    "project_id": "my-app-id"
  },
  "auth": {
    "enabled": true,
    "user_collection": "users"
  },
  "models": [
    {
      "name": "posts",
      "protected": true,
      "fields": {
        "title": "string",
        "content": "text",
        "published": "boolean"
      },
      "relations": {
        "comments": "hasMany:comments",
        "author": "belongsTo:users"
      }
    },
    {
      "name": "comments",
      "protected": true,
      "fields": {
        "text": "string",
        "timestamp": "datetime"
      },
      "relations": {
        "post_id": "belongsTo:posts"
      }
    }
  ]
}
```
