module example.com/student-service

go 1.23.0

toolchain go1.24.4

require (
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace example.com/student-service => ./.
