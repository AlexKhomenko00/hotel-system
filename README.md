# Hotel Management System

An example hotel reservation system built in Go, inspired by [ByteByteGo](https://bytebytego.com/) system design principles. Features JWT authentication, optimistic locking for reservations, and cloud-native deployment on Azure with PostgreSQL, Kubernetes, and Terraform infrastructure.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. The system includes REST APIs for hotel management, user authentication, and reservation booking with optimistic locking.

### Prerequisites
- Go 1.24+
- Docker & Docker Compose
- PostgreSQL (or use Docker)

### Setup
1. Clone the repository
2. Copy `.env.example` to `.env` and configure your database settings
3. Start PostgreSQL database:
   ```bash
   make docker-run
   ```
4. Run database migrations:
   ```bash
   make migrate-up
   ```
5. Start the application:
   ```bash
   make run
   ```

## MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```

Database migrations:
```bash
make migrate-up          # Run migrations
make migrate-down        # Rollback migrations
make migrate-status      # Check migration status
make migrate-create name="migration_name"  # Create new migration
```

Code generation:
```bash
make generate           # Generate SQLC code from SQL
```

Terraform & Azure deployment:
```bash
make tf-init           # Initialize Terraform
make tf-plan           # Plan infrastructure changes
make tf-apply          # Apply infrastructure
make tf-destroy        # Destroy infrastructure
make deploy-azure      # Deploy to Azure Kubernetes
```
