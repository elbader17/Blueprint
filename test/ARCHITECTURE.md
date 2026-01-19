# Architecture of Generated Project

This project follows **Hexagonal Architecture** (also known as Ports and Adapters) and **Clean Code** principles. It supports multiple database drivers (Firestore, PostgreSQL, MongoDB) through a unified repository interface.

## Directory Structure

```
<project_name>/
├── cmd/
│   └── api/
│       └── main.go           # Entry point: wires everything together
├── internal/
│   ├── domain/               # Core business logic (Ports)
│   │   ├── <model>.go        # Model struct and Repository interface
│   ├── infrastructure/       # External concerns (Adapters)
│   │   └── db/
│   │       ├── firestore.go  # Firestore client (if selected)
│   │       ├── postgres.go   # PostgreSQL client (if selected)
│   │       ├── mongo.go      # MongoDB client (if selected)
│   │       └── <model>_repo.go # DB-specific implementation of the Port
│   ├── handlers/             # Application Layer (Adapters)
│   │   ├── <model>/
│   │   │   ├── handler.go    # HTTP handlers for the model
│   │   │   └── handler_test.go # Unit tests for the handler
│   │   └── auth/             # Authentication handlers
│   ├── auth/                 # Auth logic and middleware
│   ├── payments/             # Payment provider integrations
│   └── config/               # Configuration management
└── ...
```

## Core Concepts

### 1. Database Abstraction

The project uses a **Repository Pattern** to abstract database operations. The domain layer defines interfaces (Ports), and the infrastructure layer provides implementations (Adapters) for the selected database:

- **Firestore**: Uses the official Google Cloud Firestore SDK.
- **PostgreSQL**: Uses `pgx` for high-performance SQL operations.
- **MongoDB**: Uses the official MongoDB Go driver.

### 2. Dependency Injection

All dependencies are injected in `cmd/api/main.go`. The database client is initialized based on the configuration and passed to the model-specific repositories.

### 3. Clean Code & Guard Clauses

The code uses **guard clauses** to keep the logic flat and readable, avoiding deep nesting.

## How to Work with This Code

1.  **Switching Databases**: To change the database, you would typically update the initialization in `main.go` and provide the corresponding repository implementation in `internal/infrastructure/db`.
2.  **Adding a Field**: Update the domain struct and the repository implementation.
3.  **Environment Variables**:
    - `DATABASE_URL`: Required for PostgreSQL and MongoDB.
    - `MOCK_AUTH`: Set to `true` to bypass Firebase Auth during development.
