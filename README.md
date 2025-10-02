# Key-Value In Memory Store 

An in memory key-value store with a REST API gateway and gRPC communication between services.

### Services

- **API Gateway**: HTTP REST server that handles client requests and forwards them to a gRPC client
- **Key-Value Service**: gRPC server that manages in memory key-value storage

## Quick Start

1. **Start both services with docker:**
   ```bash
   docker compose up --build
   ```

2. **Use the API:**
   ```bash
   # API key sent in the header
   x-api-key: my-secret-key
   
   # Health check
   curl http://localhost:8888/health

   # Set a key-value pair
   curl -X PUT http://localhost:8888/v1/values \
     -H "Content-Type: application/json" \
     -H "x-api-key: my-secret-key" \
     -d '{"key": "hello", "value": "world"}'

   # Get a value
   curl -X GET http://localhost:8888/v1/values/hello \
     -H "x-api-key: my-secret-key"

   # Delete a key
   curl -X DELETE http://localhost:8888/v1/values/hello \
     -H "x-api-key: my-secret-key"
   ```

## Run Tests
  ```bash
  go test ./...
  ```

## Development Setup Requirements

### Prerequisites

- **Go 1.24+**
- **Docker & Docker Compose**
- **Protocol Buffers Compiler** 

### Regenerate Protobuf
```bash
protoc --go_out=. --go_opt=module=key-value --go-grpc_out=. --go-grpc_opt=module=key-value proto/keyvalue.proto
```

### Key Components

- **`client/`**: A gRPC client library for connecting to the key-value service
- **`proto/`**: Protocol buffer definitions and generated code
- **`services/api-gateway/`**: HTTP REST API that proxies requests to the key-value service
- **`services/key-value/`**: Core gRPC service that manages key-value storage

## Assumptions
- All keys and values are strings.
- There is no persistance between restarts
- Only surface level security. Hardcoded secrets in docker files. No auth between interservice communication.
- Everything is commited to the repo to make delivery easier (env files, docker files with secrets, and debug configurations)

<details>
<summary>Mistakes</summary>

- Handler constructor confusion There were two and one was not using the interface. Naming - fixed
- Race condition in tests - NOT FIXED
- KVStore client not getting closed. Moved it to main and passed to the router. Defer close - fixed

</details>