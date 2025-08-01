# Shopping List API

[![CI](https://github.com/uriberma/shopping-list-api/actions/workflows/ci.yml/badge.svg)](https://github.com/uriberma/shopping-list-api/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-66.9%25-yellow)](https://github.com/uriberma/shopping-list-api/actions)
[![Go Report](https://img.shields.io/badge/go%20report-A+-brightgreen)](https://github.com/uriberma/shopping-list-api)
[![Go Version](https://img.shields.io/badge/go-1.24-blue)](https://golang.org/)

A RESTful API for managing shopping lists and items, built with Go, Gin, PostgreSQL, and following hexagonal architecture and domain-driven design principles.

## Features

- **CRUD operations** for shopping lists and items
- **Hexagonal architecture** with clean separation of concerns
- **Domain-driven design** with rich domain models
- **PostgreSQL** database with GORM
- **API versioning** (v1)
- **Docker support** for easy deployment
- **CORS enabled** for frontend integration

## Architecture

The project follows hexagonal architecture with these layers:

```
├── cmd/server/              # Application entry point
├── internal/
│   ├── domain/             # Domain layer (entities, repositories)
│   ├── application/        # Application layer (services, use cases)
│   ├── infrastructure/     # Infrastructure layer (database, persistence)
│   └── adapters/          # Adapters layer (HTTP handlers, routes)
```

## API Endpoints

### Shopping Lists

- `POST /api/v1/lists` - Create a new shopping list
- `GET /api/v1/lists` - Get all shopping lists
- `GET /api/v1/lists/{id}` - Get a specific shopping list
- `PUT /api/v1/lists/{id}` - Update a shopping list
- `DELETE /api/v1/lists/{id}` - Delete a shopping list

### Items

- `POST /api/v1/lists/{listId}/items` - Add item to shopping list
- `GET /api/v1/lists/{listId}/items` - Get all items in a shopping list
- `GET /api/v1/items/{id}` - Get a specific item
- `PUT /api/v1/items/{id}` - Update an item
- `DELETE /api/v1/items/{id}` - Delete an item
- `PATCH /api/v1/items/{id}/toggle` - Toggle item completion status

### Health Check

- `GET /health` - API health check

## Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL 15+
- Docker & Docker Compose (optional)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd shopping-list-api
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your database configuration
   ```

4. **Start PostgreSQL**
   ```bash
   docker-compose up postgres -d
   ```

5. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

### Using Docker

1. **Start all services**
   ```bash
   docker-compose up -d
   ```

The API will be available at `http://localhost:8080`

## Example Usage

### Create a Shopping List

```bash
curl -X POST http://localhost:8080/api/v1/lists \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Weekly Groceries",
    "description": "Groceries for this week"
  }'
```

### Add Items to the List

```bash
curl -X POST http://localhost:8080/api/v1/lists/{list-id}/items \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Milk",
    "quantity": 2
  }'
```

### Get All Shopping Lists

```bash
curl http://localhost:8080/api/v1/lists
```

### Toggle Item Completion

```bash
curl -X PATCH http://localhost:8080/api/v1/items/{item-id}/toggle
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL username | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `password` |
| `DB_NAME` | PostgreSQL database name | `shopping_list_db` |
| `DB_SSLMODE` | PostgreSQL SSL mode | `disable` |
| `PORT` | Server port | `8080` |
| `GIN_MODE` | Gin mode (debug/release) | `release` |

## Project Structure

```
shopping-list-api/
├── cmd/server/                           # Application entry point
│   └── main.go
├── internal/
│   ├── domain/                          # Domain layer
│   │   ├── entities/                    # Domain entities
│   │   │   ├── shopping_list.go
│   │   │   ├── item.go
│   │   │   └── errors.go
│   │   └── repositories/                # Repository interfaces
│   │       └── shopping_list_repository.go
│   ├── application/                     # Application layer
│   │   └── services/                    # Business logic services
│   │       ├── shopping_list_service.go
│   │       └── item_service.go
│   ├── infrastructure/                  # Infrastructure layer
│   │   ├── database/                    # Database configuration
│   │   │   └── postgres.go
│   │   └── persistence/                 # Repository implementations
│   │       ├── postgres_shopping_list_repository.go
│   │       └── postgres_item_repository.go
│   └── adapters/                        # Adapters layer
│       └── http/                        # HTTP adapters
│           ├── handlers/                # HTTP handlers
│           │   ├── shopping_list_handler.go
│           │   └── item_handler.go
│           └── routes/                  # Route definitions
│               └── routes.go
├── docker-compose.yml                   # Docker Compose configuration
├── Dockerfile                          # Docker image definition
├── .env.example                        # Environment variables example
├── go.mod                              # Go module definition
└── README.md                           # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License.
