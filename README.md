# blueprint.cli

A powerful tool to generate Go API projects from a simple markdown blueprint.

## Features

- 🚀 **Fast Generation**: Create a production-ready Go project structure in seconds.
- 🔒 **Integrated Authentication**: Optional support for Firebase Auth (Login, Register, Roles).
- 📦 **Automatic CRUD**: Generates handlers and routes to create, read, update, and delete documents.
- 🛡️ **Protected Routes**: Easily configure which models require authentication.
- 🧪 **Unit Tests**: Automatically generates unit tests for all endpoints.
- 💳 **Payment Integration**: Easily enable payments with **Mercado Pago** or **Stripe**.
- 🐳 **Docker Ready**: Automatically generates `Dockerfile` and `docker-compose.yml`.
- 📚 **Swagger Docs**: Automatically generates Swagger documentation for your API.
- 📄 **Simple Configuration**: Everything is defined in a single `blueprint.md` file.
- 🖥️ **Interactive TUI**: Visual wizard to create your blueprint.

## Prerequisites

- [Go 1.21+](https://go.dev/dl/) installed.
- A [Firebase](https://console.firebase.google.com/) project.

## Installation

Clone the repository and build the generator using the provided Makefile:

```bash
git clone https://github.com/elbader17/Blueprint
cd Blueprint
make build
```

This will download dependencies and build the `blueprint_gen` binary.

## Usage

### Interactive Mode (Recommended)

Run the generator without arguments to launch the interactive Terminal User Interface (TUI):

```bash
./blueprint_gen
```

The interactive wizard will guide you through:
1.  **Project Setup**: Define project name and database type (Firestore, PostgreSQL, MongoDB).
2.  **Authentication**: Enable authentication and configure user collection.
3.  **Models**: Create data models with fields and types.
4.  **Relations**: Define relationships between models (e.g., `author` -> `User`).

### Manual Mode

You can also provide an existing blueprint file directly:

```bash
./blueprint_gen blueprint.md
```

## Getting Credentials

For the generated API to work correctly, you need to configure your Firebase project.

### 1. Project ID
1. Go to the [Firebase Console](https://console.firebase.google.com/).
2. Select your project.
3. Go to **Project settings** (gear icon).
4. Copy the **Project ID** (e.g., `my-shop-123`). You will use this in your `blueprint.md` file.

### 2. Service Account Key
1. In the same **Project settings** section, go to the **Service accounts** tab.
2. Click on **Generate new private key**.
3. A JSON file will be downloaded.
4. **Rename** this file to `firebaseCredentials.json`.
5. **Place it** in the root of the directory where you will run the generator (or in the root of your generated API).

> ⚠️ **IMPORTANT**: Never upload this file to a public repository. Add it to your `.gitignore`.

### Auth Options

#### Firebase Auth (Default)
```json
"auth": {
  "enabled": true,
  "provider": "firebase",
  "user_collection": "users"
}
```

#### JWT Auth (Simple & Robust)
```json
"auth": {
  "enabled": true,
  "provider": "jwt",
  "user_collection": "users"
}
```
If using `jwt`, the following endpoints are added:
- `POST /auth/register`: Create a new user with email/password.
- `POST /auth/login`: Login and receive a JWT token.

Set `JWT_SECRET` in your environment variables.
If you enable the `payments` module, you need credentials for your chosen provider.

#### Mercado Pago
1. Go to the [Mercado Pago Developers](https://www.mercadopago.com.ar/developers/panel) panel.
2. Select your application (or create a new one).
3. Go to **Production Credentials** or **Test Credentials**.
4. Copy the **Access Token**.
5. Set it as an environment variable: `export MP_ACCESS_TOKEN=your_token_here`

#### Stripe
1. Go to the [Stripe Dashboard](https://dashboard.stripe.com/apikeys).
2. Copy the **Secret Key**.
3. Create a Webhook Endpoint and copy the **Webhook Secret**.
4. Set them as environment variables:
   ```bash
   export STRIPE_SECRET_KEY=your_secret_key_here
   export STRIPE_WEBHOOK_SECRET=your_webhook_secret_here
   ```

### Docker Support

The tool automatically generates a `Dockerfile` and `docker-compose.yml`. You can start your API and database with a single command:

```bash
docker-compose up -d --build
```

This will:
1. Build your Go API.
2. Start the database (Postgres, MongoDB).
3. Connect them together.

For a detailed explanation of the architecture, directory structure, and how to work with the generated code, see the **`ARCHITECTURE.md`** file inside your generated project.

Key architectural features:
- **Domain Layer**: Core models and repository interfaces (Ports).
- **Infrastructure Layer**: Firestore implementations (Adapters).
- **Application Layer**: HTTP handlers (Adapters) and dependency injection.
- **Guard Clauses**: Clean, readable code without nested `if-else`.
- **Modular Handlers**: Model-specific logic organized in dedicated folders.

## Blueprint Configuration Guide

The `blueprint.md` file is the heart of your project. It defines your API's architecture, database, authentication, and business logic in a single, human-readable file.

### 1. Creating the File

You can create a `blueprint.md` file in two ways:

**Option A: Interactive Wizard (Recommended)**
Run the tool without arguments. It will ask you questions and generate the file for you.
```bash
./blueprint_gen
```

**Option B: Manual Creation**
Create a new file named `blueprint.md` in your project root and add the JSON configuration inside a markdown code block.

### 2. JSON Structure Reference

The `blueprint.md` file must contain a single JSON code block. Here is the complete schema:

```json
{
  "project_name": "string (Required)",
  "database": { ... },
  "auth": { ... },
  "payments": { ... },
  "models": [ ... ]
}
```

#### Global Settings
| Field | Type | Description |
|-------|------|-------------|
| `project_name` | `string` | The name of your project. This will be used for the module name (go.mod) and the root folder. |

#### Database Configuration (`database`)
Configures the data storage layer.

**Firestore (Firebase)**
```json
"database": {
  "type": "firestore",
  "project_id": "your-firebase-project-id"
}
```

**PostgreSQL**
```json
"database": {
  "type": "postgresql",
  "url": "postgres://user:pass@localhost:5432/dbname"
}
```

**MongoDB**
```json
"database": {
  "type": "mongodb",
  "url": "mongodb://localhost:27017"
}
```

#### Authentication (`auth`)
(Optional) Enables built-in JWT authentication and user management.

```json
"auth": {
  "enabled": true,
  "user_collection": "users"
}
```
- `enabled`: Set to `true` to generate auth endpoints (`/login`, `/register`).
- `user_collection`: The name of the database table/collection to store user data.

#### Payments (`payments`)
(Optional) Integrates payment processing.

```json
"payments": {
  "enabled": true,
  "provider": "mercadopago", // or "stripe"
  "transactions_collection": "transactions"
}
```
- `provider`: Supports `mercadopago` or `stripe`.
- `transactions_collection`: Where to store payment logs.

#### Data Models (`models`)
Defines your application's entities (tables/collections).

```json
{
  "name": "products",
  "protected": false,
  "fields": {
    "name": "string",
    "price": "float",
    "description": "text",
    "is_active": "boolean",
    "created_at": "datetime"
  },
  "relations": {
    "category": "belongsTo:categories",
    "reviews": "hasMany:reviews"
  }
}
```

- **`name`**: The name of the resource (plural recommended, e.g., "products").
- **`protected`**: If `true`, all endpoints for this model will require an Authorization header.
- **`fields`**: A key-value map defining the data structure.
    - Supported types:
        - `string`: Short text (e.g., name, email).
        - `text`: Long text (e.g., description, bio).
        - `integer`: Whole numbers (e.g., quantity, age).
        - `float`: Decimal numbers (e.g., price, rating).
        - `boolean`: True/False flags.
        - `datetime`: Date and time timestamps.
- **`relations`**: Defines how models connect to each other.
    - `belongsTo:<model_name>`: Many-to-One relationship (e.g., A product belongs to a category).
    - `hasMany:<model_name>`: One-to-Many relationship (e.g., A user has many orders).

### 3. Full Example

Copy this into your `blueprint.md` to get started:

````markdown
# My API Blueprint

```json
{
  "project_name": "MyAwesomeAPI",
  "database": {
    "type": "firestore",
    "project_id": "my-app-id"
  },
  "auth": {
    "enabled": true,
    "user_collection": "users"
  },
  "models": [
    {
      "name": "posts",
      "protected": true,
      "fields": {
        "title": "string",
        "content": "text",
        "published": "boolean"
      },
      "relations": {
        "comments": "hasMany:comments",
        "author": "belongsTo:users"
      }
    },
    {
      "name": "comments",
      "protected": true,
      "fields": {
        "text": "string",
        "timestamp": "datetime"
      },
      "relations": {
        "post_id": "belongsTo:posts"
      }
    }
  ]
}
```
````

## Generated Endpoints

If you enable `auth`, the following will be available:

- `POST /auth/login`: Login.
- `POST /auth/register`: Register a new user.
- `GET /auth/me`: Get current user profile (Requires Token).
- `GET /auth/roles`: List available roles (Requires Token).

If you enable `payments` with **Mercado Pago**:

- `POST /payments/mercadopago/preference`: Create a payment preference.
- `POST /payments/mercadopago/webhook`: Webhook to receive payment notifications.

If you enable `payments` with **Stripe**:

- `POST /payments/stripe/payment-intent`: Create a PaymentIntent.
- `POST /payments/stripe/webhook`: Webhook to receive payment events.

> [!IMPORTANT]
> To use Mercado Pago, set `MP_ACCESS_TOKEN`.
> To use Stripe, set `STRIPE_SECRET_KEY` and `STRIPE_WEBHOOK_SECRET`.

For each model (e.g., `products`):

- `GET /api/products`: List all.
- `GET /api/products/:id`: Get one.
- `POST /api/products`: Create one.
- `PUT /api/products/:id`: Update one.
- `DELETE /api/products/:id`: Delete one.

If the model is `protected: true`, you must send the header:
`Authorization: Bearer <FIREBASE_ID_TOKEN>`
