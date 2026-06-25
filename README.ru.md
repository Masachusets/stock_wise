# StockWise

Система учёта оборудования и материальных средств для управления активами, закреплениями и перемещениями.

## Технологический стек

- **Язык**: Go 1.26+
- **API-фреймворк**: [Goa v3](https://goa.design) — design-first, генерация кода
- **База данных**: PostgreSQL 16 (через `jackc/pgx/v5`)
- **Логирование**: `log/slog` (структурированные JSON-логи)
- **Импорт Excel**: `xuri/excelize/v2`
- **Инфраструктура**: Docker Compose (только PostgreSQL)

## Структура проекта

```
cmd/server/               # Точка входа HTTP-сервера
cmd/import/               # CLI-утилита: массовый импорт из Excel в PostgreSQL
internal/
  config/                 # Конфигурация (флаги, переменные окружения)
  nomenclatures/          # Справочник номенклатур (только чтение)
  departments/            # Справочник подразделений (только чтение)
  cards/                  # Карточки сотрудников (CRUD)
  equipments/             # Оборудование (CRUD) с информацией о закреплении
  waybills/               # Накладные (CRUD) + подписание/архивирование
  assignments/            # История закреплений (только чтение)
  importer/               # Логика импорта из Excel
design/design.go          # Goa-дизайн API (источник истины)
sql/create.sql            # Схема БД (DDL, ENUM, индексы)
gen/                      # [gitignored] Сгенерированный Goa-код
```

## Быстрый старт

### Требования
- Go 1.26+
- Docker & Docker Compose
- PostgreSQL 16 (через Docker)

### Запуск

```bash
# 1. Запустить PostgreSQL
docker compose up -d

# 2. Импортировать данные из Excel
go run ./cmd/import -db "postgres://stockwise:stockwise@localhost:5432/stockwise" -excel ./excel

# 3. Запустить API-сервер
go run ./cmd/server -db "postgres://stockwise:stockwise@localhost:5432/stockwise" -port 8080
```

### Флаги сервера

| Флаг   | По умолчанию                                             | Описание               |
|--------|----------------------------------------------------------|------------------------|
| `-db`  | `postgres://stockwise:stockwise@localhost:5432/stockwise` | Подключение к PostgreSQL |
| `-port`| `8080`                                                   | HTTP-порт              |
| `-log` | `info`                                                   | Уровень логирования (debug/info/warn/error) |

## API-эндпоинты

| Метод  | Путь                                | Описание                           |
|--------|-------------------------------------|------------------------------------|
| GET    | /dictionaries/nomenclatures         | Список номенклатур                 |
| GET    | /dictionaries/nomenclatures/{id}    | Номенклатура по ID                 |
| GET    | /dictionaries/departments           | Список подразделений               |
| GET    | /dictionaries/departments/{code}    | Подразделение по коду              |
| GET    | /cards                              | Список карточек                    |
| GET    | /cards/{number}                     | Карточка по номеру                 |
| POST   | /cards                              | Создать карточку                   |
| PUT    | /cards/{number}                     | Обновить карточку                  |
| DELETE | /cards/{number}                     | Удалить карточку                   |
| GET    | /equipments                         | Список оборудования (с фильтрами) |
| GET    | /equipments/{inventory_number}      | Оборудование с закреплением        |
| POST   | /equipments                         | Создать оборудование               |
| PUT    | /equipments/{inventory_number}      | Обновить оборудование              |
| DELETE | /equipments/{inventory_number}      | Удалить оборудование (мягкое)      |
| GET    | /waybills                           | Список накладных                   |
| GET    | /waybills/{id}                      | Накладная с позициями              |
| POST   | /waybills                           | Создать накладную                  |
| POST   | /waybills/{id}/sign                 | Подписать (DRAFT→SIGNED)           |
| POST   | /waybills/{id}/archive              | Архивировать (SIGNED→ARCHIVED)     |
| DELETE | /waybills/{id}                      | Удалить накладную (только DRAFT)   |
| GET    | /assignments                        | Список закреплений                 |
| GET    | /assignments/{id}                   | Закрепление по ID                  |

## Схема БД

Основные таблицы:
- `nomenclatures` — справочник номенклатур (код + наименование)
- `cards` — карточки сотрудников (номер карточки как PK)
- `departments` — подразделения (код INT как PK, тип enum)
- `equipments` — оборудование (инвентарный номер `ИТ\d{5}`, мягкое удаление)
- `waybills` — накладные (draft → signed → archived)
- `waybills_equipments` — связь накладных и оборудования
- `equipments_assignments` — история закреплений оборудования

## Перегенерация кода

После изменения `design/design.go`:

```bash
goa gen github.com/Masachusets/stock_wise/design
go build ./cmd/server
```
