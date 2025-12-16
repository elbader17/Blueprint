# Blueprint Generator

Blueprint is a CLI tool written in Go that automatically generates a complete RESTful API using the **Gin Framework** and **Firestore**. It allows you to define your database structure and authentication configuration in a simple Markdown file, and the tool takes care of creating all the necessary code.

## Features

- 🚀 **Fast Generation**: Create a production-ready Go project structure in seconds.
- 🔒 **Integrated Authentication**: Optional support for Firebase Auth (Login, Register, Roles).
- 📦 **Automatic CRUD**: Generates handlers and routes to create, read, update, and delete documents.
- 🛡️ **Protected Routes**: Easily configure which models require authentication.
- 🧪 **Unit Tests**: Automatically generates unit tests for all endpoints.
- 📚 **Swagger Docs**: Automatically generates Swagger documentation for your API.
- 📄 **Simple Configuration**: Everything is defined in a single `blueprint.md` file.

## Prerequisites

- [Go 1.21+](https://go.dev/dl/) installed.
- A [Firebase](https://console.firebase.google.com/) project.

## Installation

Clone the repository and build the generator:

```bash
git clone https://github.com/elbader17/Blueprint
cd Blueprint
go mod tidy
go build -o blueprint_gen cmd/blueprint/main.go
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

## Usage

1. Create a `blueprint.md` file with your API definition (see example below).
2. Run the generator passing the file as an argument:

```bash
./blueprint_gen blueprint.md
```

3. The generator will:
   - Create the project structure.
   - Install dependencies.
   - Generate Swagger documentation.
   - **Automatically start the API server**.

4. You will see the server logs in your terminal. You can stop it with `Ctrl+C`.

## Manual Run

If you want to run the API manually later:

```bash
cd ShopMasterAPI
go run cmd/api/main.go
```

## Testing

The generator automatically creates unit tests for your endpoints. To run them:

```bash
cd ShopMasterAPI
go test ./...
```

## Documentation (Swagger)

The API comes with auto-generated Swagger documentation.

1. Run the API.
2. Open your browser and navigate to:
   `http://localhost:8080/swagger/index.html`

To update the documentation after making changes to the code (if you modify the generated code manually):

```bash
swag init -g cmd/api/main.go
```

## Blueprint Format

The input file must contain a JSON code block with the following structure:

### Complete Example (`blueprint.md`)

```markdown
# My E-Commerce Project

System architecture definition.

` + "```" + `json
{
  "project_name": "ShopMasterAPI",
  "firestore_project_id": "shop-master-prod",
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
        "description": "text",
        "in_stock": "boolean"
      },
      "relations": {
        "category": "belongsTo:categories"
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
        "items": "hasMany:order_items"
      }
    }
  ]
}
` + "```" + `
```

### Field Explanation

- **`project_name`**: Name of the folder and Go module that will be generated.
- **`firestore_project_id`**: Your Firebase Project ID (obtained in the credentials step).
- **`auth`** (Optional):
    - `enabled`: `true` to activate the login/register system.
    - `user_collection`: Name of the Firestore collection where users will be stored (e.g., "users").
- **`models`**: List of your database entities.
    - `name`: Name of the collection in Firestore.
    - `protected`: If `true`, routes for this model will require a valid Bearer token.
    - `fields`: Map of `field_name: type`. Supported types: `string`, `text`, `integer`, `float`, `boolean`, `datetime`.
    - `relations`: (Informational) Defines relations between models.

## Generated Endpoints

If you enable `auth`, the following will be available:

- `POST /auth/login`: Login.
- `POST /auth/register`: Register a new user.
- `GET /auth/me`: Get current user profile (Requires Token).
- `GET /auth/roles`: List available roles (Requires Token).

For each model (e.g., `products`):

- `GET /api/products`: List all.
- `GET /api/products/:id`: Get one.
- `POST /api/products`: Create one.
- `PUT /api/products/:id`: Update one.
- `DELETE /api/products/:id`: Delete one.

If the model is `protected: true`, you must send the header:
`Authorization: Bearer <FIREBASE_ID_TOKEN>`
