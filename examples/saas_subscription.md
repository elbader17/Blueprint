# SaaS Subscription Platform

This blueprint creates a backend for a SaaS platform with **Stripe** payments.

## Blueprint

```json
{
  "project_name": "SaaSBackend",
  "database": {
    "type": "postgresql",
    "url": "postgres://user:pass@localhost:5432/saas_db"
  },
  "auth": {
    "enabled": true,
    "user_collection": "users"
  },
  "payments": {
    "enabled": true,
    "provider": "stripe",
    "transactions_collection": "payments"
  },
  "models": [
    {
      "name": "subscriptions",
      "protected": true,
      "fields": {
        "plan_id": "string",
        "status": "string",
        "start_date": "datetime",
        "end_date": "datetime"
      },
      "relations": {
        "user_id": "belongsTo:users"
      }
    },
    {
      "name": "plans",
      "protected": false,
      "fields": {
        "name": "string",
        "price": "float",
        "interval": "string",
        "stripe_price_id": "string"
      },
      "relations": {}
    }
  ]
}
```
