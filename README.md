# student-service

Простое gRPC-приложение на Go для управления данными студентов.

## Структура проекта
```text
student-service/
├── go.mod
├── main.go
├── proto/
│   ├── student.proto
│   ├── student.pb.go
│   └── student_grpc.pb.go
├── server/
│   └── service.go
```

## Установка зависимостей

```bash
go mod tidy
Генерация gRPC-кода
Перед генерацией убедитесь, что установлены необходимые плагины:

bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
Добавьте ~/go/bin в PATH, если это не сделано:

bash
export PATH="$PATH:$(go env GOPATH)/bin"
Сгенерируйте .pb.go файлы:

bash
protoc \
  --go_out=paths=source_relative:. \
  --go-grpc_out=paths=source_relative:. \
  proto/student.proto
Запуск сервера

go run main.go

gRPC-сервер будет слушать порт :50051.

Включение рефлексии

Для возможности тестирования с grpcurl необходимо подключить рефлексию в main.go:

import "google.golang.org/grpc/reflection"

...

reflection.Register(grpcServer)

Также необходимо установить зависимость:

go get google.golang.org/grpc/reflection

Тестирование с grpcurl

Убедитесь, что grpcurl установлен:https://github.com/fullstorydev/grpcurl

Примеры запросов:

CreateStudent

grpcurl -plaintext -d '{
  "first_name": "Ivan",
  "last_name": "Petrov",
  "grade": 9
}' localhost:50051 student.StudentService/CreateStudent

GetStudent

grpcurl -plaintext -d '{
  "id": "PASTE_ID_HERE"
}' localhost:50051 student.StudentService/GetStudent

UpdateStudent

grpcurl -plaintext -d '{
  "student": {
    "id": "PASTE_ID_HERE",
    "first_name": "Ivan",
    "last_name": "Ivanov",
    "grade": 10
  }
}' localhost:50051 student.StudentService/UpdateStudent

DeleteStudent

grpcurl -plaintext -d '{
  "id": "PASTE_ID_HERE"
}' localhost:50051 student.StudentService/DeleteStudent

Зависимости

Go 1.20+

gRPC

protoc

protoc-gen-go

protoc-gen-go-grpc

grpcurl (для тестов)