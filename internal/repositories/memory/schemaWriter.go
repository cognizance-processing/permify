package memory

import (
	"context"
	"errors"

	"permify/internal/repositories"
	db "permify/pkg/database/memory"
	"permify/pkg/logger"
	base "permify/pkg/pb/base/v1"
)

// SchemaWriter - Structure for Schema Writer
type SchemaWriter struct {
	database *db.Memory
	// logger
	logger logger.Interface
}

// NewSchemaWriter creates a new SchemaWriter
func NewSchemaWriter(database *db.Memory, logger logger.Interface) *SchemaWriter {
	return &SchemaWriter{
		database: database,
		logger:   logger,
	}
}

// WriteSchema - Write Schema to repository
func (w *SchemaWriter) WriteSchema(ctx context.Context, definitions []repositories.SchemaDefinition) error {
	var err error
	txn := w.database.DB.Txn(true)
	defer txn.Abort()
	for _, definition := range definitions {
		if err = txn.Insert(SchemaDefinitionsTable, definition); err != nil {
			return errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}
	txn.Commit()
	return nil
}
