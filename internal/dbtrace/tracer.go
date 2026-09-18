// Package dbtrace logs every SQL statement the pgx pool runs. It is opt-in
// (SQL_LOG=true in the environment, see internal/app) and meant for
// development: it's how you verify a change actually reduced the number of
// queries - e.g. the Project.files dataloader turning N queries into 1.
package dbtrace

import (
	"context"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Tracer implements pgx.QueryTracer.
type Tracer struct{}

var _ pgx.QueryTracer = Tracer{}

func (Tracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	log.Printf("SQL %s", label(data.SQL))
	return ctx
}

func (Tracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {}

// label prefers the sqlc query name (every generated query starts with
// "-- name: GetFilesByProjectIDs :many") and otherwise falls back to the
// first line of the statement.
func label(sql string) string {
	sql = strings.TrimSpace(sql)
	first, _, _ := strings.Cut(sql, "\n")
	if rest, ok := strings.CutPrefix(first, "-- name: "); ok {
		name, _, _ := strings.Cut(rest, " ")
		return name
	}
	if len(first) > 80 {
		first = first[:80] + "..."
	}
	return first
}
