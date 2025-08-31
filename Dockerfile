# syntax=docker/dockerfile:1

# Go version
ARG GO_VERSION=1.23
ARG BIN_NAME=student-service

FROM golang:${GO_VERSION} AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build from ./cmd, because that's where main.go lives
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w -extldflags '-static'" \
    -o /out/${BIN_NAME} ./cmd

FROM gcr.io/distroless/static:nonroot
WORKDIR /app

COPY --from=build /out/${BIN_NAME} /app/${BIN_NAME}
COPY config.yaml /app/config.yaml
ENV CONFIG_PATH=/app/config.yaml

EXPOSE 50051
USER nonroot:nonroot
ENTRYPOINT ["/app/student-service"]
