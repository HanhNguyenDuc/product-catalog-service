# Product Catalog Service

This service provides product management capabilities with an event-driven architecture.

## Design decicions
- I follow the hexagonal architecture pattern, split the persistence layer from the business logic layer (domain layer).
- The Domain layer is trying to design follow the way that domain expert follow, also my domain layer has things called change tracking, to help reduce the cost of query make in the database (persistence layer)
- Usecase layer include lot of things, i don't put repository implementation directly as usecase attribute, instead of that, i create usecase necessary interface, the main reason i do that is to make the usecase layer more testable, and also make it easier to switch between different database implementation, usecase also doesn't need to care about persistence layer.
- I use Dependency Injection to manage the dependencies between layers, i use wire to generate the dependency injection code, the Dependency Injection framework I use is uber/fx, that gonna help me to reduce cost of initializing the dependencies.
- This projects also has 2 ports, one is grpc port, another is rest port, the grpc port is the one that i have try end to end test, the rest port is just the implementation to help people who doesn't familiar with grpc.
- I use Cloud Spanner as the database, and I use the emulator to run it locally.

## 🛠 Local Setup (Spanner Emulator)

Use the provided scripts or commands to set up the Google Cloud Spanner emulator locally.

### 0) Start Spanner Emulator

```bash
docker compose up -d
```

### 1) Create & activate an emulator config

```bash
gcloud config configurations create spanner-emulator
gcloud config configurations activate spanner-emulator
```

### 2) Point gcloud at the emulator (REST endpoint) & disable auth

```bash
gcloud config set auth/disable_credentials true
gcloud config set project emulator-project
gcloud config set api_endpoint_overrides/spanner http://localhost:9020/
```

### 3) Create instance & database

```bash
gcloud spanner instances create test-instance \
  --config=emulator-config \
  --description="Local Spanner Emulator" \
  --nodes=1

gcloud spanner databases create test-db --instance=test-instance
```

### 4) Apply schema (DDL) - Migrations

I usually use GORM for migration with golang codebase, I'm kind of new to spanner so to make it not be complicated, I'm using raw sql file to apply schema.

```bash
gcloud spanner databases ddl update test-db \
  --instance=test-instance \
  --ddl-file=migrations/001_initial_schema.sql
```

---

## 🚀 Running the Service

Start the gRPC server:

```bash
go run cmd/server/main.go
```

---

## Testing

The service includes a CLI tool (`cmd/client/main.go`) to test operations locally.

### Create a product

```bash
go run cmd/client/main.go create --name test --desc testdesc --cat testcategory
```

### Update a product

```bash
go run cmd/client/main.go update --id <product-id> --name updatedName --desc desc123123123
```

### Get a product

```bash
go run cmd/client/main.go get --id <product-id>
```

### List products

```bash
go run cmd/client/main.go list
```

### Activate a product

```bash
go run cmd/client/main.go activate --id <product-id>
```

### Deactivate a product

```bash
go run cmd/client/main.go deactivate --id <product-id>
```

### Apply a discount

```bash
go run cmd/client/main.go discount apply --id <product-id> --pct 50 --duration 24h
```

### Remove a discount

```bash
go run cmd/client/main.go discount remove --id <product-id>
```

---

## 🔍 Database Inspection

You can use `gcloud` to run queries against the local Spanner emulator:

### Check product data

```bash
gcloud spanner databases execute-sql test-db \
  --instance=test-instance \
  --sql="SELECT * FROM products"
```

### Check outbox events data

```bash
gcloud spanner databases execute-sql test-db \
  --instance=test-instance \
  --sql="SELECT * FROM outbox_events"
```