package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel/codes"

	"permify/internal/repositories"
	"permify/internal/schema"
	db "permify/pkg/database/postgres"
	"permify/pkg/logger"
	base "permify/pkg/pb/base/v1"
)

// SchemaReader - Structure for SchemaReader
type SchemaReader struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
	// logger
	logger logger.Interface
}

// NewSchemaReader - Creates a new SchemaReader
func NewSchemaReader(database *db.Postgres, logger logger.Interface) *SchemaReader {
	return &SchemaReader{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true},
		logger:    logger,
	}
}

// ReadSchema - Reads entity config from the repository.
func (r *SchemaReader) ReadSchema(ctx context.Context, tenantID, version string) (sch *base.SchemaDefinition, err error) {
	ctx, span := tracer.Start(ctx, "schema-reader.read-schema")
	defer span.End()

	builder := r.database.Builder.Select("entity_type, serialized_definition, version").From(SchemaDefinitionTable).Where(squirrel.Eq{"version": version, "tenant_id": tenantID})

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	var rows *sql.Rows
	rows, err = r.database.DB.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()

	var definitions []string
	for rows.Next() {
		sd := repositories.SchemaDefinition{}
		err = rows.Scan(&sd.EntityType, &sd.SerializedDefinition, &sd.Version)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		definitions = append(definitions, sd.Serialized())
	}
	if err = rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	sch, err = schema.NewSchemaFromStringDefinitions(true, definitions...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return sch, err
}

// ReadSchemaDefinition - Reads entity config from the repository.
func (r *SchemaReader) ReadSchemaDefinition(ctx context.Context, tenantID, entityType, version string) (definition *base.EntityDefinition, v string, err error) {
	ctx, span := tracer.Start(ctx, "schema-reader.read-schema-definition")
	defer span.End()

	builder := r.database.Builder.Select("entity_type, serialized_definition, version").Where(squirrel.Eq{"entity_type": entityType, "version": version, "tenant_id": tenantID}).From(SchemaDefinitionTable).Limit(1)

	var query string
	var args []interface{}

	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}

	var def repositories.SchemaDefinition
	row := r.database.DB.QueryRowContext(ctx, query, args...)
	if err = row.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	if err = row.Scan(&def.EntityType, &def.SerializedDefinition, &def.Version); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
		}
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCAN.String())
	}

	var sch *base.SchemaDefinition
	sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", err
	}

	definition, err = schema.GetEntityByName(sch, entityType)
	return definition, def.Version, err
}

// HeadVersion - Finds the latest version of the schema.
func (r *SchemaReader) HeadVersion(ctx context.Context, tenantID string) (version string, err error) {
	ctx, span := tracer.Start(ctx, "schema-reader.head-version")
	defer span.End()

	var query string
	var args []interface{}
	query, args, err = r.database.Builder.
		Select("version").From(SchemaDefinitionTable).Where(squirrel.Eq{"tenant_id": tenantID}).OrderBy("version DESC").Limit(1).
		ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}
	row := r.database.DB.QueryRowContext(ctx, query, args...)
	err = row.Scan(&version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
		}
		return "", err
	}

	return version, nil
}
