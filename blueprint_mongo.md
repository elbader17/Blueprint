# Mongo Project Blueprint

```json
{
  "project_name": "MongoAPI",
  "database": {
    "type": "mongodb",
    "url": "mongodb://localhost:27017"
  },
  "auth": {
    "enabled": true,
    "user_collection": "users"
  },
  "models": [
    {
      "name": "products",
      "protected": false,
      "fields": {
        "name": "string",
        "price": "float",
        "description": "text"
      }
    }
  ]
}
```
