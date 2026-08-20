# Appointment Booking System (Turnero)

This blueprint defines a multi-tenant appointment management system. It uses JWT authentication and MongoDB.

```json
{
  "project_name": "turnero-api",
  "database": {
    "type": "mongodb",
    "url": "mongodb://user:password@localhost:27019"
  },
  "auth": {
    "enabled": true,
    "provider": "jwt",
    "user_collection": "users"
  },
  "payments": {
    "enabled": false
  },
  "models": [
    {
      "name": "accounts",
      "fields": {
        "name": "string",
        "subscription_status": "string",
        "created_at": "datetime"
      },
      "relations": {
        "users": "hasMany:users"
      }
    },
    {
      "name": "users",
      "protected": true,
      "fields": {
        "account_id": "string",
        "role": "string"
      }
    },
    {
      "name": "locations",
      "protected": true,
      "fields": {
        "account_id": "string",
        "name": "string",
        "address": "string",
        "phone": "string"
      }
    },
    {
      "name": "professionals",
      "protected": true,
      "fields": {
        "account_id": "string",
        "location_id": "string",
        "name": "string",
        "specialty": "string",
        "active": "boolean"
      }
    },
    {
      "name": "services",
      "protected": true,
      "fields": {
        "account_id": "string",
        "name": "string",
        "description": "text",
        "duration_minutes": "integer",
        "price": "float"
      }
    },
    {
      "name": "appointments",
      "protected": true,
      "fields": {
        "account_id": "string",
        "professional_id": "string",
        "location_id": "string",
        "service_id": "string",
        "customer_name": "string",
        "customer_phone": "string",
        "start_time": "datetime",
        "status": "string"
      }
    }
  ]
}
```

### Key Features:
- **Multi-user per Account**: All entities (locations, professionals, etc.) are tied to an `account_id`.
- **JWT Auth**: Secured endpoints using standard tokens.
- **Flexible Management**: Separate models for Locations and Professionals to support multi-branch businesses.
- **MongoDB**: Ready for high-volume document storage.
