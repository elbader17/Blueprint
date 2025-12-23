# Social Network Blueprint

A social platform with posts, likes, and groups.

```json
{
  "project_name": "SocialNetworkAPI",
  "database": {
    "type": "firestore",
    "project_id": "social-network-demo"
  },
  "auth": {
    "enabled": true,
    "user_collection": "users"
  },
  "models": [
    {
      "name": "users",
      "protected": true,
      "fields": {
        "username": "string",
        "bio": "string",
        "verified": "boolean"
      },
      "relations": {
        "posts": "hasMany:posts",
        "comments": "hasMany:comments",
        "groups_owned": "hasMany:groups"
      }
    },
    {
      "name": "posts",
      "protected": true,
      "fields": {
        "content": "string",
        "image_url": "string",
        "likes_count": "integer",
        "created_at": "datetime"
      },
      "relations": {
        "author_id": "belongsTo:users",
        "comments": "hasMany:comments",
        "likes": "hasMany:likes",
        "group_id": "belongsTo:groups"
      }
    },
    {
      "name": "comments",
      "protected": true,
      "fields": {
        "text": "string",
        "created_at": "datetime"
      },
      "relations": {
        "post_id": "belongsTo:posts",
        "author_id": "belongsTo:users"
      }
    },
    {
      "name": "likes",
      "protected": true,
      "fields": {
        "created_at": "datetime"
      },
      "relations": {
        "post_id": "belongsTo:posts",
        "user_id": "belongsTo:users"
      }
    },
    {
      "name": "groups",
      "protected": true,
      "fields": {
        "name": "string",
        "description": "string",
        "is_private": "boolean"
      },
      "relations": {
        "admin_id": "belongsTo:users",
        "posts": "hasMany:posts"
      }
    }
  ]
}
```
