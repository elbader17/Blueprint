# test Blueprint

```json
{
  "project_name": "test",
  "database": {
    "type": "firestore",
    "project_id": "your-project-id"
  },
  "auth": {
    "enabled": true,
    "user_collection": "User"
  },
  "models": [
    {
      "fields": {
        "created_at": "datetime",
        "name": "string"
      },
      "name": "account",
      "protected": true,
      "relations": {
        "user": "hasMany:User"
      }
    }
  ]
}
```
