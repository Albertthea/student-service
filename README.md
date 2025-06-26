# student-service

A simple gRPC-based service written in Go for managing student data.

## Project Structure
```text
student-service/
├── cmd/         # Application entry point(s)
├── proto/       # Protobuf definitions
├── service/      # Business logic and gRPC service implementation
├── go.mod       # Go module definition
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

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Run the linter:
```bash
golangci-lint run
```
GitHub Actions Linting also runs on every push and pull request to main via GitHub Actions.

Workflow config: .github/workflows/lint.yml

## Requirements
Go 1.20+

Protocol Buffers compiler (protoc) >= 3.21.0

gRPC codegen plugins
  protoc-gen-go v1.30.0+
  protoc-gen-go-grpc v1.3.0+
