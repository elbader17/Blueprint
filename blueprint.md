# E-Commerce Platform Blueprint

This blueprint defines a comprehensive e-commerce system with authentication, product management, order processing, and user reviews.

```json
{
  "project_name": "ShopMasterAPI",
  "firestore_project_id": "test-blueprint-c2632",
  "auth": {
    "enabled": true,
    "user_collection": "users"
  },
  "models": [
    {
      "name": "categories",
      "protected": false,
      "fields": {
        "name": "string",
        "slug": "string",
        "description": "text",
        "image_url": "string",
        "active": "boolean"
      },
      "relations": {
        "products": "hasMany"
      }
    },
    {
      "name": "products",
      "protected": false,
      "fields": {
        "name": "string",
        "sku": "string",
        "description": "text",
        "price": "float",
        "stock_quantity": "integer",
        "images": "string", 
        "is_featured": "boolean",
        "created_at": "datetime"
      },
      "relations": {
        "category_id": "belongsTo:categories",
        "reviews": "hasMany"
      }
    },
    {
      "name": "reviews",
      "protected": true,
      "fields": {
        "rating": "integer",
        "comment": "text",
        "created_at": "datetime"
      },
      "relations": {
        "user_id": "belongsTo:users",
        "product_id": "belongsTo:products"
      }
    },
    {
      "name": "carts",
      "protected": true,
      "fields": {
        "updated_at": "datetime"
      },
      "relations": {
        "user_id": "belongsTo:users",
        "items": "hasMany:cart_items"
      }
    },
    {
      "name": "cart_items",
      "protected": true,
      "fields": {
        "quantity": "integer"
      },
      "relations": {
        "cart_id": "belongsTo:carts",
        "product_id": "belongsTo:products"
      }
    },
    {
      "name": "orders",
      "protected": true,
      "fields": {
        "order_number": "string",
        "total_amount": "float",
        "status": "string",
        "payment_status": "string",
        "shipping_address": "text",
        "created_at": "datetime"
      },
      "relations": {
        "user_id": "belongsTo:users",
        "items": "hasMany:order_items"
      }
    },
    {
      "name": "order_items",
      "protected": true,
      "fields": {
        "quantity": "integer",
        "price_at_purchase": "float"
      },
      "relations": {
        "order_id": "belongsTo:orders",
        "product_id": "belongsTo:products"
      }
    },
    {
      "name": "wishlists",
      "protected": true,
      "fields": {
        "created_at": "datetime"
      },
      "relations": {
        "user_id": "belongsTo:users",
        "products": "hasMany:products"
      }
    }
  ]
}
```
