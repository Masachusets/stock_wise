# StockWise

Equipment and inventory tracking system for managing material resources, fixed assets, and assignments.

## Tech Stack

- **Language**: Go 1.26+
- **API Framework**: [Goa v3](https://goa.design) — design-first, code-generated
- **Database**: PostgreSQL 16 (via `jackc/pgx/v5`)
- **Logging**: `log/slog` (structured JSON logs)
- **Excel Import**: `xuri/excelize/v2`
- **Dev Infra**: Docker Compose (PostgreSQL only)

## Project Structure

```
cmd/server/               # HTTP server entry point
cmd/import/               # CLI tool: bulk import from Excel to PostgreSQL
internal/
  config/                 # Configuration (flags, env)
  nomenclatures/          # Read-only dictionary service
  departments/            # Read-only dictionary service
  cards/                  # Staff cards CRUD
  equipments/             # Equipment CRUD with assignment info
  waybills/               # Waybills CRUD + sign/archive lifecycle
  assignments/            # Assignment history (read-only)
  importer/               # Excel import logic
design/design.go          # Goa API design (source of truth)
sql/create.sql            # Database schema (DDL, ENUMs, indexes)
gen/                      # [gitignored] Goa-generated code
```

## Quick Start

### Prerequisites
- Go 1.26+
- Docker & Docker Compose
- PostgreSQL 16 (via Docker)

### Run

```bash
# 1. Start PostgreSQL
docker compose up -d

# 2. Import data from Excel
go run ./cmd/import -db "postgres://stockwise:stockwise@localhost:5432/stockwise" -excel ./excel

# 3. Start the API server
go run ./cmd/server -db "postgres://stockwise:stockwise@localhost:5432/stockwise" -port 8080
```

### Server Flags

| Flag   | Default                                                  | Description            |
|--------|----------------------------------------------------------|------------------------|
| `-db`  | `postgres://stockwise:stockwise@localhost:5432/stockwise` | PostgreSQL connection  |
| `-port`| `8080`                                                   | HTTP port              |
| `-log` | `info`                                                   | Log level (debug/info/warn/error) |

## API Endpoints

| Method | Path                                | Description                    |
|--------|-------------------------------------|--------------------------------|
| GET    | /dictionaries/nomenclatures         | List nomenclatures             |
| GET    | /dictionaries/nomenclatures/{id}    | Get nomenclature by ID         |
| GET    | /dictionaries/departments           | List departments               |
| GET    | /dictionaries/departments/{code}    | Get department by code         |
| GET    | /cards                              | List staff cards               |
| GET    | /cards/{number}                     | Get card by number             |
| POST   | /cards                              | Create card                    |
| PUT    | /cards/{number}                     | Update card                    |
| DELETE | /cards/{number}                     | Delete card                    |
| GET    | /equipments                         | List equipments (with filters) |
| GET    | /equipments/{inventory_number}      | Get equipment with assignment  |
| POST   | /equipments                         | Create equipment               |
| PUT    | /equipments/{inventory_number}      | Update equipment               |
| DELETE | /equipments/{inventory_number}      | Delete equipment (soft)        |
| GET    | /waybills                           | List waybills                  |
| GET    | /waybills/{id}                      | Get waybill with items         |
| POST   | /waybills                           | Create waybill                 |
| POST   | /waybills/{id}/sign                 | Sign waybill (DRAFT→SIGNED)    |
| POST   | /waybills/{id}/archive              | Archive waybill (SIGNED→ARCHIVED) |
| DELETE | /waybills/{id}                      | Delete waybill (DRAFT only)    |
| GET    | /assignments                        | List assignments               |
| GET    | /assignments/{id}                   | Get assignment                 |

## Database Schema

Core tables:
- `nomenclatures` — product catalog (code + name)
- `cards` — staff (card_number as PK)
- `departments` — organizational units (code INT as PK, type enum)
- `equipments` — tracked equipment (inventory_number `ИТ\d{5}`, soft delete)
- `waybills` — transfer documents (draft → signed → archived)
- `waybills_equipments` — waybill-to-equipment binding
- `equipments_assignments` — equipment-to-employee/department binding history

## Regenerate Code

After changing `design/design.go`:

```bash
goa gen github.com/Masachusets/stock_wise/design
go build ./cmd/server
```
