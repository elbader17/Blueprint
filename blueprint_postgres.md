# Postgres Project Blueprint

```json
{
  "project_name": "PostgresAPI",
  "database": {
    "type": "postgresql",
    "url": "postgres://user:password@localhost:5432/dbname"
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
