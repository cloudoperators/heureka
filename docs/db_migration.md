# Database Migration Process

This document describes the database migration process for the Heureka application, using the [`golang-migrate`](https://github.com/golang-migrate/migrate) tool. Migration files are versioned SQL scripts and are executed automatically at the application startup to ensure the database schema is always up-to-date.

---

## Migration File Structure

All migration files are stored in:

```
internal/database/mariadb/migrations/
```

Each migration consists of two SQL files:

- `X_description.up.sql` — defines the changes to **apply** to the database.
- `X_description.down.sql` — defines how to **revert** those changes.

Where:
- `X` is a migration version number (e.g., `001`, `002`, etc.). Higher number is newer version. In Heureka creation date 'YYYYMMDDhhmmss' is used as script version.
- `description` is a descriptive name of the migration (e.g., `add_users_table`).

Example:
```
20250406154423_add_users_table.up.sql
20250406154423_add_users_table.down.sql
```

---

## Embedding Migrations

The migration files are embedded into the Go binary using Go's [`embed`](https://pkg.go.dev/embed) package. This ensures all necessary files are bundled with the application binary at build time.

---

## Migration Execution at Startup

The Heureka application runs the migration process automatically at startup. It uses the `golang-migrate` library with a MariaDB driver to:

- Load the embedded SQL migration files.
- Apply any **pending migrations** to bring the database schema to the **latest version**.

This ensures consistency and eliminates manual intervention when deploying new versions of the app.

---

## Migration Versioning

Each pair of migration files is versioned sequentially:

- Versions are increasing integers (e.g., `001`, `002`, `003`). In Heureka creation date 'YYYYMMDDhhmmss' is used as script version.
- New migrations **must** follow the next available number in sequence.
- The migration tool tracks applied versions in a special schema migrations table.

---

## Makefile Targets

The project includes helpful `Makefile` targets to simplify migration tasks.

### `install-migrate`

Installs the `golang-migrate` CLI tool:

```bash
make install-migrate
```

This will install the `migrate` binary locally for use in creating and managing migrations.

### `create-migration`

Creates a new migration file pair. Requires a `MIGRATION_NAME` environment variable:

```bash
make create-migration MIGRATION_NAME=add_products_table
```

This command will generate:

```
internal/database/mariadb/migrations/XXX_add_products_table.up.sql
internal/database/mariadb/migrations/XXX_add_products_table.down.sql
```

Where `XXX` is the next migration version number, created using date schema 'YYYYMMDDhhmmss'

---

## Example Migration Files

### `20250408202133_add_users_table.up.sql`

```sql
CREATE TABLE users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### `20250408202133_add_users_table.down.sql`

```sql
DROP TABLE IF EXISTS users;
```

---

## 📄 `schema_migrations` Table (Migration State Tracking)

When using [golang-migrate](https://github.com/golang-migrate/migrate), the tool automatically manages migration state through a table called `schema_migrations`.

### 🔍 Purpose

This table keeps track of the current migration version applied to the database. It ensures that:
- Migrations are applied in order.
- No migration is re-applied.
- Failed migrations are flagged to prevent partial application.

### 🧱 Table Structure

```sql
CREATE TABLE schema_migrations (
  version bigint NOT NULL,
  dirty boolean NOT NULL
);
```

- **`version`**: The version number of the last successfully applied migration (based on the filename prefix, e.g., `001_init.up.sql` → version `1`).
- **`dirty`**: Set to `true` if the last migration did not complete successfully. This must be resolved before further migrations can be applied.

### ⚠️ Notes

- This table is automatically created by `golang-migrate` upon first run.
- Do **not** manually modify this table unless you're intentionally correcting a migration state (e.g., rolling back a failed migration).
- `golang-migrate` does **not** verify the contents of migration files using checksums (like MD5); it solely tracks versions via filenames and this metadata table.

### ✅ Best Practices

- Always version your migration files clearly using the numeric prefix (e.g., `001_`, `002_`, etc.).
- Commit your migration files to version control to prevent tampering or accidental modification.
- Use the `dirty` flag as an indicator in CI/CD to detect failed or partial deployments.

---

## References

- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [Go `embed` package](https://pkg.go.dev/embed)
- [MariaDB Documentation](https://mariadb.com/kb/en/documentation/)

---

## ✅ Summary

- Migrations are embedded and executed on app startup.
- Use `make create-migration MIGRATION_NAME=...` to create new migration files.
- Migrations are stored in `internal/database/mariadb/migrations`.
- Versioning and execution handled by `golang-migrate`.
