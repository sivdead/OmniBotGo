# Database migrations

Migrations are **dialect-specific** (E02). golang-migrate loads one of:

| `db.type` | Source directory |
|---|---|
| `mysql` (default) | `migrations/mysql` |
| `postgres` / `postgresql` | `migrations/postgres` |
| `sqlite` / `sqlite3` | `migrations/sqlite` |

## Verified in Batch 4 agent env

- **SQLite**: migrate Up smoke + FK pragma DSN helper unit tests
- **MySQL / Postgres**: URL/source selection unit-tested; live migrate smoke only if those servers are available in the environment

## Notes

- MySQL keeps the historical BIGINT → UUID path (`20250805021718_*`). UUID **down** is refused (E06).
- Postgres/SQLite start with VARCHAR/TEXT UUID IDs and include `updated_at` on log tables from `000001`.
- MySQL additive `20261021000000_add_log_updated_at` adds `updated_at` to existing log tables (E07).
