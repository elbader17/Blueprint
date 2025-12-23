# Project Management Blueprint

A Jira-like tool for managing projects, tasks, and boards.

```json
{
  "project_name": "ProjectManagerAPI",
  "database": {
    "type": "mongodb",
    "url": "mongodb://user:password@localhost:27017"
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
        "email": "string",
        "avatar_url": "string"
      },
      "relations": {
        "owned_projects": "hasMany:projects",
        "assigned_tasks": "hasMany:tasks",
        "comments": "hasMany:comments"
      }
    },
    {
      "name": "projects",
      "protected": true,
      "fields": {
        "name": "string",
        "description": "string",
        "start_date": "datetime",
        "end_date": "datetime"
      },
      "relations": {
        "owner_id": "belongsTo:users",
        "boards": "hasMany:boards"
      }
    },
    {
      "name": "boards",
      "protected": true,
      "fields": {
        "name": "string",
        "type": "string"
      },
      "relations": {
        "project_id": "belongsTo:projects",
        "columns": "hasMany:columns"
      }
    },
    {
      "name": "columns",
      "protected": true,
      "fields": {
        "name": "string",
        "order": "integer"
      },
      "relations": {
        "board_id": "belongsTo:boards",
        "tasks": "hasMany:tasks"
      }
    },
    {
      "name": "tasks",
      "protected": true,
      "fields": {
        "title": "string",
        "description": "string",
        "priority": "string",
        "due_date": "datetime"
      },
      "relations": {
        "column_id": "belongsTo:columns",
        "assignee_id": "belongsTo:users",
        "comments": "hasMany:comments"
      }
    },
    {
      "name": "comments",
      "protected": true,
      "fields": {
        "content": "string",
        "created_at": "datetime"
      },
      "relations": {
        "task_id": "belongsTo:tasks",
        "author_id": "belongsTo:users"
      }
    }
  ]
}
```
