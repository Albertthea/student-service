# student-service

A simple gRPC-based service written in Go for managing student data.  
Uses PostgreSQL for persistent storage, running via Docker for local development and testing.

## Project Structure
```text
student-service/
├── cmd/                        # Application entry point(s)
├── proto/                      # Protobuf definitions
├── repository/student/         # DB model and repository logic
├── repository/migrations/      # SQL migrations
├── service/                    # Business logic and gRPC service implementation
├── docker-compose.yml          # PostgreSQL container setup
├── go.mod / go.sum             # Go module definition
```

## Code Generation

Ensure `$GOPATH/bin` is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Then generate the gRPC code:

```bash
protoc \
  --go_out=paths=source_relative:. \
  --go-grpc_out=paths=source_relative:. \
  proto/student.proto
```

## Running the Server

```bash
go run cmd/main.go
```

The server listens on port `:50051`.

## Run PostgreSQL via Docker
```bash
docker-compose up -d
```

It starts a postgres:15 container with:
User: student
Password: 111111
Database: studentdb
Port: 5432

## Run DB Migration
After the container is up, apply the initial schema:

```bash
psql -h localhost -U student -d studentdb -f repository/migrations/00000_initial.sql
```
Password: 111111

## Environment Variables for DB Credentials

The application expects the database login and password to be provided via environment variables.

The names of these variables are configured in `config.yaml` under the PostgreSQL authorization environment section:

```yaml
postgresql:
  authorisation:
    env:
      login: DB_USER
      password: DB_PASS
```

Set these environment variables before running the server:

```bash
export DB_USER=student
export DB_PASS=111111
```

Make sure these variables match what you specify in your config.yaml.
This allows the application to securely load the DB credentials without hardcoding them in the code or config files.

## Testing with grpcurl

Make sure grpcurl is installed.

Example Requests:

### CreateStudent

```bash
grpcurl -plaintext -d '{
  "first_name": "Ivan",
  "last_name": "Petrov",
  "grade": 9
}' localhost:50051 student.StudentService/CreateStudent
```

### GetStudent

```bash
grpcurl -plaintext -d '{
  "id": "PASTE_ID_HERE"
}' localhost:50051 student.StudentService/GetStudent
```

### UpdateStudent

```bash
grpcurl -plaintext -d '{
  "student": {
    "id": "PASTE_ID_HERE",
    "first_name": "Ivan",
    "last_name": "Ivanov",
    "grade": 10
  }
}' localhost:50051 student.StudentService/UpdateStudent
```

### DeleteStudent

```bash
grpcurl -plaintext -d '{
  "id": "PASTE_ID_HERE"
}' localhost:50051 student.StudentService/DeleteStudent
```

### Linting
Using golangci-lint for static analysis.


Local Run Install golangci-lint:

Install golangci-lint:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Run the linter:

```bash
golangci-lint run
```

GitHub Actions
Linting also runs on every push and pull request to main via GitHub Actions.

Workflow config: .github/workflows/lint.yml

## Requirements
Go 1.20+
Docker + docker-compose

Protocol Buffers compiler (protoc) >= 3.21.0

gRPC codegen plugins
  protoc-gen-go v1.30.0+
  protoc-gen-go-grpc v1.3.0+
