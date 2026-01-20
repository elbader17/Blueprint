# Enterprise ERP Blueprint

A comprehensive Enterprise Resource Planning (ERP) system blueprint including inventory management, supply chain, e-commerce orders, support tickets, and extensive audit logging.

This example demonstrates a large-scale application structure with 17 interconnected models, authentication, and payments.

```json
{
  "project_name": "EnterpriseERP",
  "database": {
    "type": "postgresql",
    "url": "postgres://user:pass@localhost:5432/enterprise_db"
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
        "full_name": "string",
        "phone": "string",
        "address": "text",
        "is_verified": "boolean",
        "last_login": "datetime"
      },
      "relations": {
        "orders": "hasMany:orders",
        "tickets": "hasMany:support_tickets",
        "reviews": "hasMany:reviews",
        "wishlists": "hasMany:wishlists",
        "audit_logs": "hasMany:audit_logs"
      }
    },
    {
      "name": "transactions",
      "protected": true,
      "fields": {
        "amount": "float",
        "status": "string",
        "provider_id": "string",
        "currency": "string",
        "created_at": "datetime"
      },
      "relations": {
        "order": "belongsTo:orders"
      }
    },
    {
      "name": "products",
      "protected": false,
      "fields": {
        "sku": "string",
        "name": "string",
        "description": "text",
        "price": "float",
        "cost_price": "float",
        "weight": "float",
        "dimensions": "string",
        "is_active": "boolean",
        "created_at": "datetime"
      },
      "relations": {
        "category": "belongsTo:categories",
        "supplier": "belongsTo:suppliers",
        "inventory": "hasMany:inventory_stocks",
        "reviews": "hasMany:reviews",
        "order_items": "hasMany:order_items"
      }
    },
    {
      "name": "categories",
      "protected": false,
      "fields": {
        "name": "string",
        "slug": "string",
        "description": "text",
        "is_visible": "boolean"
      },
      "relations": {
        "products": "hasMany:products"
      }
    },
    {
      "name": "suppliers",
      "protected": true,
      "fields": {
        "company_name": "string",
        "contact_name": "string",
        "email": "string",
        "phone": "string",
        "tax_id": "string",
        "contract_end": "datetime"
      },
      "relations": {
        "products": "hasMany:products",
        "purchase_orders": "hasMany:purchase_orders"
      }
    },
    {
      "name": "warehouses",
      "protected": true,
      "fields": {
        "name": "string",
        "location_code": "string",
        "address": "text",
        "capacity": "integer",
        "manager_name": "string"
      },
      "relations": {
        "stocks": "hasMany:inventory_stocks"
      }
    },
    {
      "name": "inventory_stocks",
      "protected": true,
      "fields": {
        "quantity": "integer",
        "aisle": "string",
        "bin": "string",
        "last_audited": "datetime",
        "restock_threshold": "integer"
      },
      "relations": {
        "product": "belongsTo:products",
        "warehouse": "belongsTo:warehouses"
      }
    },
    {
      "name": "orders",
      "protected": true,
      "fields": {
        "order_number": "string",
        "total_amount": "float",
        "tax_amount": "float",
        "status": "string",
        "shipping_address": "text",
        "placed_at": "datetime"
      },
      "relations": {
        "customer": "belongsTo:users",
        "items": "hasMany:order_items",
        "payment": "belongsTo:transactions",
        "shipment": "hasMany:shipments",
        "invoice": "hasMany:invoices",
        "coupon": "belongsTo:coupons"
      }
    },
    {
      "name": "order_items",
      "protected": true,
      "fields": {
        "quantity": "integer",
        "unit_price": "float",
        "discount": "float",
        "total": "float"
      },
      "relations": {
        "order": "belongsTo:orders",
        "product": "belongsTo:products"
      }
    },
    {
      "name": "shipments",
      "protected": true,
      "fields": {
        "tracking_number": "string",
        "carrier": "string",
        "shipped_at": "datetime",
        "estimated_arrival": "datetime",
        "status": "string",
        "weight": "float"
      },
      "relations": {
        "order": "belongsTo:orders"
      }
    },
    {
      "name": "invoices",
      "protected": true,
      "fields": {
        "invoice_number": "string",
        "issued_at": "datetime",
        "due_date": "datetime",
        "total": "float",
        "status": "string",
        "pdf_url": "string"
      },
      "relations": {
        "order": "belongsTo:orders"
      }
    },
    {
      "name": "support_tickets",
      "protected": true,
      "fields": {
        "subject": "string",
        "message": "text",
        "priority": "string",
        "status": "string",
        "created_at": "datetime",
        "closed_at": "datetime"
      },
      "relations": {
        "user": "belongsTo:users"
      }
    },
    {
      "name": "reviews",
      "protected": true,
      "fields": {
        "rating": "integer",
        "comment": "text",
        "approved": "boolean",
        "created_at": "datetime",
        "likes": "integer"
      },
      "relations": {
        "user": "belongsTo:users",
        "product": "belongsTo:products"
      }
    },
    {
      "name": "wishlists",
      "protected": true,
      "fields": {
        "name": "string",
        "created_at": "datetime",
        "is_public": "boolean"
      },
      "relations": {
        "user": "belongsTo:users",
        "items": "hasMany:wishlist_items"
      }
    },
    {
      "name": "wishlist_items",
      "protected": true,
      "fields": {
        "added_at": "datetime",
        "priority": "string"
      },
      "relations": {
        "wishlist": "belongsTo:wishlists",
        "product": "belongsTo:products"
      }
    },
    {
      "name": "coupons",
      "protected": true,
      "fields": {
        "code": "string",
        "discount_percent": "float",
        "expires_at": "datetime",
        "active": "boolean",
        "usage_limit": "integer"
      },
      "relations": {
        "orders": "hasMany:orders"
      }
    },
    {
      "name": "audit_logs",
      "protected": true,
      "fields": {
        "action": "string",
        "resource": "string",
        "resource_id": "string",
        "timestamp": "datetime",
        "ip_address": "string",
        "details": "text"
      },
      "relations": {
        "user": "belongsTo:users"
      }
    }
  ]
}
```
