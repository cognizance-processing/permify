package utils

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/Masterminds/squirrel"

	"permify/pkg/logger"
)

// SnapshotQuery -
func SnapshotQuery(sl squirrel.SelectBuilder, revision uint64) squirrel.SelectBuilder {
	return sl.Where(squirrel.Or{
		squirrel.Expr(fmt.Sprintf("pg_visible_in_snapshot(created_tx_id, (select snapshot from transactions where id = '%v'::xid8)) = true", revision)),
		squirrel.Expr(fmt.Sprintf("created_tx_id = '%v'::xid8", revision)),
	}).Where(squirrel.And{
		squirrel.Or{
			squirrel.Expr(fmt.Sprintf("pg_visible_in_snapshot(expired_tx_id, (select snapshot from transactions where id = '%v'::xid8)) = false", revision)),
			squirrel.Expr("expired_tx_id = '0'::xid8"),
		},
		squirrel.Expr(fmt.Sprintf("expired_tx_id <> '%v'::xid8", revision)),
	})
}

// GarbageCollectQuery -
func GarbageCollectQuery(window time.Duration, tenantID string) squirrel.DeleteBuilder {
	return squirrel.Delete("relation_tuples").
		Where(squirrel.Expr(fmt.Sprintf("created_tx_id IN (SELECT id FROM transactions WHERE timestamp < '%v')", time.Now().Add(-window).Format(time.RFC3339)))).
		Where(squirrel.And{
			squirrel.Or{
				squirrel.Expr("expired_tx_id = '0'::xid8"),
				squirrel.Expr(fmt.Sprintf("expired_tx_id IN (SELECT id FROM transactions WHERE timestamp < '%v')", time.Now().Add(-window).Format(time.RFC3339))),
			},
			squirrel.Expr(fmt.Sprintf("tenant_id = '%v'", tenantID))})

}

// Rollback - Rollbacks a transaction and logs the error
func Rollback(tx *sql.Tx, logger logger.Interface) {
	if err := tx.Rollback(); !errors.Is(err, sql.ErrTxDone) && err != nil {
		logger.Error("failed to rollback transaction", err)
	}
}
