# Go Migration Patterns

Canonical patterns for database migrations, store types, and DB connection. Uses goose for migrations.

## Migration file format

```sql
-- migrations/NNN_description.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE items (
    id          BIGSERIAL    PRIMARY KEY,
    name        TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
-- +goose StatementEnd
```

Rules:
- Always number sequentially: `001`, `002`, `003` — never skip or duplicate.
- Before creating a new migration: `Glob("migrations/**")` and find the highest NNN. Next = NNN + 1.
- Always include both `-- +goose Up` and `-- +goose Down`.
- Always wrap statements in `StatementBegin` / `StatementEnd`.
- Use `IF NOT EXISTS` in Up, `IF EXISTS` in Down.

## DB connection

```go
// internal/db/db.go
package db

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
)

// Connect opens and verifies a PostgreSQL connection.
// DSN: postgres://user:pass@host:port/dbname?sslmode=disable
func Connect(dsn string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("db: open: %w", err)
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("db: ping: %w", err)
    }
    return db, nil
}
```

## Migration runner

```go
// internal/db/migrate.go
package db

import (
    "database/sql"
    "fmt"
    "github.com/pressly/goose/v3"
)

// RunMigrations applies all pending goose migrations from migrationsDir.
func RunMigrations(db *sql.DB, migrationsDir string) error {
    goose.SetDialect("postgres")
    if err := goose.Up(db, migrationsDir); err != nil {
        return fmt.Errorf("migrations: %w", err)
    }
    return nil
}
```

## Store pattern

```go
// internal/db/<entity>_store.go
package db

import (
    "database/sql"
    "errors"
    "fmt"
)

var ErrNotFound = errors.New("item: not found")

const (
    queryCreate  = `INSERT INTO items (name) VALUES ($1) RETURNING id, name, created_at`
    queryGetByID = `SELECT id, name, created_at FROM items WHERE id = $1`
    queryList    = `SELECT id, name, created_at FROM items ORDER BY created_at DESC`
)

type ItemStore struct{ db *sql.DB }

func NewItemStore(db *sql.DB) *ItemStore { return &ItemStore{db: db} }

func (s *ItemStore) Create(ctx context.Context, name string) (Item, error) {
    var item Item
    err := s.db.QueryRowContext(ctx, queryCreate, name).Scan(&item.ID, &item.Name, &item.CreatedAt)
    if err != nil {
        return Item{}, fmt.Errorf("itemstore.Create: %w", err)
    }
    return item, nil
}

func (s *ItemStore) GetByID(ctx context.Context, id int64) (Item, error) {
    var item Item
    err := s.db.QueryRowContext(ctx, queryGetByID, id).Scan(&item.ID, &item.Name, &item.CreatedAt)
    if errors.Is(err, sql.ErrNoRows) {
        return Item{}, ErrNotFound
    }
    if err != nil {
        return Item{}, fmt.Errorf("itemstore.GetByID: %w", err)
    }
    return item, nil
}
```

Rules:
- SQL as package-level `const` strings — never inline SQL in method bodies.
- Return `ErrNotFound` (not `sql.ErrNoRows`) from store methods — callers must not depend on DB internals.
- Always pass `context.Context` as first arg to every store method.

## Store test pattern

```go
// internal/db/<entity>_store_test.go
func testDB(t *testing.T) *sql.DB {
    t.Helper()
    dsn := os.Getenv("TEST_DSN")
    if dsn == "" {
        t.Skip("TEST_DSN not set")
    }
    db, err := Connect(dsn)
    if err != nil {
        t.Fatalf("testDB: %v", err)
    }
    if err := RunMigrations(db, "../../migrations"); err != nil {
        t.Fatalf("testDB migrations: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}

func TestItemStore_Create(t *testing.T) {
    if testing.Short() { t.Skip() }
    db := testDB(t)
    store := NewItemStore(db)
    item, err := store.Create(context.Background(), "test")
    if err != nil { t.Fatal(err) }
    if item.Name != "test" { t.Errorf("want test, got %s", item.Name) }
}
```
