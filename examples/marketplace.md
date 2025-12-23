# Marketplace Blueprint

A complex e-commerce platform with users, products, orders, and reviews.

```json
{
  "project_name": "MarketplaceAPI",
  "database": {
    "type": "postgresql",
    "url": "postgres://user:password@localhost:5432/marketplace"
  },
  "auth": {
    "enabled": true,
    "user_collection": "users"
  },
  "payments": {
    "enabled": true,
    "provider": "mercadopago",
    "transactions_collection": "transactions"
  },
  "models": [
    {
      "name": "users",
      "protected": true,
      "fields": {
        "name": "string",
        "email": "string",
        "role": "string"
      },
      "relations": {
        "products": "hasMany:products",
        "orders": "hasMany:orders",
        "reviews": "hasMany:reviews"
      }
    },
    {
      "name": "products",
      "protected": false,
      "fields": {
        "name": "string",
        "description": "string",
        "price": "float",
        "stock": "integer",
        "category_id": "string"
      },
      "relations": {
        "seller_id": "belongsTo:users",
        "category_id": "belongsTo:categories",
        "reviews": "hasMany:reviews",
        "order_items": "hasMany:order_items"
      }
    },
    {
      "name": "categories",
      "protected": false,
      "fields": {
        "name": "string",
        "description": "string"
      },
      "relations": {
        "products": "hasMany:products"
      }
    },
    {
      "name": "orders",
      "protected": true,
      "fields": {
        "total": "float",
        "status": "string",
        "created_at": "datetime"
      },
      "relations": {
        "user_id": "belongsTo:users",
        "items": "hasMany:order_items",
        "transaction_id": "belongsTo:transactions"
      }
    },
    {
      "name": "order_items",
      "protected": true,
      "fields": {
        "quantity": "integer",
        "price": "float"
      },
      "relations": {
        "order_id": "belongsTo:orders",
        "product_id": "belongsTo:products"
      }
    },
    {
      "name": "reviews",
      "protected": true,
      "fields": {
        "rating": "integer",
        "comment": "string",
        "created_at": "datetime"
      },
      "relations": {
        "user_id": "belongsTo:users",
        "product_id": "belongsTo:products"
      }
    }
  ]
}
```
