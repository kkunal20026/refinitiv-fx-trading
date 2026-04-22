# Refinitiv FX Trading Service

A Go-based service for fetching real-time foreign exchange (FX) rates from the Refinitiv API and booking FX deals. The service stores all rates and deals in a PostgreSQL database using Docker and exposes a REST API for clients to interact with.

---

## What This Service Does

- Fetches live FX rates (e.g. EUR to USD) from the Refinitiv API
- Falls back to built-in mock rates when the Refinitiv API is unreachable (useful for local development)
- Allows clients to book FX deals and track their status
- Stores all data in a PostgreSQL database
- Logs every request with a unique trace ID for easy debugging

---

## Project Structure

```text
refinitiv-fx-trading/
├── cmd/server/              # Starting point of the application
│   ├── main.go              # Starts the HTTP server
│   └── version.go           # Holds version information
│
├── internal/                # Core application code
│   ├── config/              # Reads settings from config.yaml or environment variables
│   ├── models/              # Data shapes (Rate, Deal, Request/Response)
│   ├── handler/             # Handles incoming HTTP requests
│   ├── service/             # Business rules (validates, processes data)
│   ├── repository/          # Reads and writes data to the database
│   ├── client/              # Connects to the Refinitiv external API
│   └── database/            # Sets up database tables on startup
│
├── pkg/                     # Shared utilities
│   ├── errors/              # Custom error types with HTTP status codes
│   ├── logger/              # Structured JSON logging
│   └── middleware/          # Request logging, CORS, panic recovery
│
├── tests/
│   ├── unit/                # Tests for individual functions
│   └── integration/         # Tests that use a real database
│
├── docker/                  # Docker build file
├── migrations/              # SQL scripts to create database tables
├── .github/workflows/       # Automated CI/CD pipeline
├── docker-compose.yml       # Starts PostgreSQL in Docker for local development
├── config.yaml              # Default configuration values
├── .env.example             # Template for environment variable overrides
├── Makefile                 # Shortcut commands for build, test, run
└── go.mod                   # Go dependency definitions
```

---

## Prerequisites

| Requirement | Version | Purpose |
| --- | --- | --- |
| [Go](https://go.dev/dl/) | 1.21 or later | Compile and run the application |
| [Docker Desktop](https://www.docker.com/products/docker-desktop/) | Latest | Run PostgreSQL locally |
| [Postman](https://www.postman.com/) or curl | Any | Test the API endpoints |

---

## Setup and Running

### Step 1 — Start the Database

Open a terminal in the project folder and run:

```bash
docker compose up -d postgres
```

This starts a PostgreSQL database container. Wait about 5 seconds for it to be ready.

### Step 2 — Download Dependencies

```bash
go mod tidy
```

### Step 3 — Start the Server

```bash
go run ./cmd/server/...
```

The server is ready when you see this in the terminal:

```text
"msg":"Server starting","address":":8080"
```

### Step 4 — Test the API

Open Postman or a browser and call:

```text
GET http://localhost:8080/health
```

Expected response:

```json
{
  "success": true,
  "data": {
    "status": "healthy"
  }
}
```

---

## Configuration

Settings are loaded from `config.yaml` automatically. To override any setting, set the matching environment variable before starting the server.

| Setting | Default | Description |
| --- | --- | --- |
| `SERVER_PORT` | `8080` | Port the HTTP server listens on |
| `DATABASE_HOST` | `localhost` | PostgreSQL host |
| `DATABASE_PORT` | `5432` | PostgreSQL port |
| `DATABASE_USER` | `postgres` | Database username |
| `DATABASE_PASSWORD` | `postgres` | Database password |
| `DATABASE_DBNAME` | `refinitiv_db` | Database name |
| `REFINITIV_BASE_URL` | `https://api.refinitiv.com` | Refinitiv API base URL |
| `REFINITIV_USERNAME` | _(empty)_ | Refinitiv account username |
| `REFINITIV_PASSWORD` | _(empty)_ | Refinitiv account password |
| `LOGGER_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |

> **Note:** When `REFINITIV_USERNAME` and `REFINITIV_PASSWORD` are not set, the service automatically returns built-in mock rates so local development works without a real Refinitiv subscription.

---

## API Endpoints

### Health Check

```text
GET /health
```

Returns the status of the service and its dependencies.

---

### Get FX Rate

```text
GET /api/v1/rates?from=EUR&to=USD
```

| Parameter | Required | Example | Description |
| --- | --- | --- | --- |
| `from` | Yes | `EUR` | Source currency (3-letter code) |
| `to` | Yes | `USD` | Target currency (3-letter code) |

**Example response:**

```json
{
  "success": true,
  "data": {
    "from_currency": "EUR",
    "to_currency": "USD",
    "bid": 1.0845,
    "ask": 1.0855,
    "mid": 1.085,
    "source": "MOCK"
  }
}
```

---

### Book a Deal

```http
POST /api/v1/deals
Content-Type: application/json
```

**Request body:**

```json
{
  "client_id": "CLIENT-001",
  "from_currency": "EUR",
  "to_currency": "USD",
  "amount": 100000,
  "direction": "BUY",
  "value_date": "2026-04-25T00:00:00Z",
  "reference": "REF-001"
}
```

| Field | Required | Allowed values | Description |
| --- | --- | --- | --- |
| `client_id` | Yes | Any string | Unique identifier for the client |
| `from_currency` | Yes | 3-letter code e.g. `EUR` | Currency being sold |
| `to_currency` | Yes | 3-letter code e.g. `USD` | Currency being bought |
| `amount` | Yes | Number greater than 0 | Trade amount |
| `direction` | Yes | `BUY` or `SELL` | Trade direction |
| `value_date` | Yes | ISO 8601 date-time | Date the trade settles |
| `reference` | No | Any string | Optional reference label |

---

### List Deals for a Client

```text
GET /api/v1/deals?client_id=CLIENT-001
```

| Parameter | Required | Description |
| --- | --- | --- |
| `client_id` | Yes | The client whose deals to list |
| `page` | No | Page number (default: 1) |
| `limit` | No | Results per page, max 100 (default: 20) |

---

### Get a Single Deal

```text
GET /api/v1/deals/{id}
```

Replace `{id}` with the UUID returned when the deal was booked.

---

### Get Deal by Trade ID

```text
GET /api/v1/deals/trade/{trade_id}
```

---

### Update Deal Status

```text
PUT /api/v1/deals/{id}/status?status=CONFIRMED
```

| Parameter | Allowed values |
| --- | --- |
| `status` | `PENDING`, `CONFIRMED`, `SETTLED`, `CANCELLED`, `REJECTED` |

---

## Error Responses

All errors follow the same format:

```json
{
  "success": false,
  "error": {
    "type": "VALIDATION_ERROR",
    "message": "Missing query parameters",
    "status_code": 400,
    "details": {
      "from": "required",
      "to": "required"
    }
  }
}
```

| Error type | HTTP status | Meaning |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | Missing or invalid input |
| `NOT_FOUND_ERROR` | 404 | Record does not exist |
| `CONFLICT_ERROR` | 409 | Duplicate record |
| `UNAUTHORIZED_ERROR` | 401 | Authentication required |
| `SERVICE_ERROR` | 502/404 | External API (Refinitiv) returned an error |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## All Commands Reference

| Command | Description |
| --- | --- |
| `docker compose up -d postgres` | Start the PostgreSQL database container in the background |
| `docker compose down` | Stop and remove the Docker containers |
| `docker compose ps` | Check which containers are currently running |
| `go mod tidy` | Download missing dependencies and regenerate `go.sum` |
| `go mod download` | Download all dependencies listed in `go.mod` |
| `go build ./...` | Compile all packages to check for errors |
| `go run ./cmd/server/...` | Compile and start the server (recommended way) |
| `go test ./...` | Run all tests |
| `go test ./tests/unit/...` | Run unit tests only |
| `go test ./tests/integration/...` | Run integration tests only |
| `go vet ./...` | Check code for common mistakes |
| `go fmt ./...` | Auto-format all Go source files |
| `make build` | Build the server binary via Makefile |
| `make run` | Run the server via Makefile |
| `make test` | Run all tests via Makefile |
| `make docker-build` | Build the Docker image for the server |
| `make docker-run` | Start all services (server + database) via Docker Compose |
| `make docker-down` | Stop all Docker services |
| `make fmt` | Format code via Makefile |
| `make lint` | Run the linter via Makefile |
| `make vet` | Run go vet via Makefile |

---

## Stopping the Server

If the server was started from the VSCode terminal, press **Ctrl+C** in that terminal window.

If the server was started via the Code Runner extension, click the **stop button (⬛)** in the Output panel, or open the Command Palette (`Ctrl+Shift+P`) and run **Stop Code Run**.

---

## Database Schema

### `rates` table

Stores every FX rate fetched from Refinitiv (or mock).

| Column | Type | Description |
| --- | --- | --- |
| `id` | UUID | Unique record identifier |
| `from_currency` | TEXT | Source currency code |
| `to_currency` | TEXT | Target currency code |
| `bid` | DECIMAL | Bid price |
| `ask` | DECIMAL | Ask price |
| `mid` | DECIMAL | Mid price |
| `timestamp` | TIMESTAMPTZ | When the rate was fetched |
| `source` | TEXT | Data source (`REFINITIV` or `MOCK`) |

### `deals` table

Stores every booked FX deal.

| Column | Type | Description |
| --- | --- | --- |
| `id` | UUID | Unique deal identifier |
| `client_id` | TEXT | Client who booked the deal |
| `trade_id` | TEXT | Unique trade reference (auto-generated) |
| `from_currency` | TEXT | Currency sold |
| `to_currency` | TEXT | Currency bought |
| `amount` | DECIMAL | Trade amount |
| `rate` | DECIMAL | Rate applied at booking |
| `status` | TEXT | Current status of the deal |
| `direction` | TEXT | `BUY` or `SELL` |
| `value_date` | TIMESTAMPTZ | Settlement date |
| `reference` | TEXT | Optional client reference |

---

## Testing

```bash
# Run all tests
go test ./...

# Run only unit tests
go test ./tests/unit/...

# Run with verbose output to see each test name
go test -v ./...
```

---

## License

MIT License — see [LICENSE](LICENSE) for details.
