package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type dbtx interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// Repository persists PostgreSQL-owned compatibility events and responses.
// A transaction callback receives a repository bound to the same transaction.
//
// The type stays handwritten because every GALA struct declaration emits Copy,
// Equal, Unapply, and StructMeta helpers that import the GALA runtime, which
// would change the Go-facing surface of this exported type. The rest of the
// former repository.go lives in repository.gala and repository_methods.go.
type Repository struct {
	db dbtx
}
