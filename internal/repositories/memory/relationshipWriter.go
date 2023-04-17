package memory

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/go-memdb"

	"permify/internal/repositories"
	"permify/internal/repositories/memory/snapshot"
	"permify/internal/repositories/memory/utils"
	"permify/pkg/database"
	db "permify/pkg/database/memory"
	"permify/pkg/logger"
	base "permify/pkg/pb/base/v1"
	"permify/pkg/token"
)

type RelationshipWriter struct {
	database *db.Memory
	// logger
	logger logger.Interface
}

// NewRelationshipWriter - Creates a new RelationshipReader
func NewRelationshipWriter(database *db.Memory, logger logger.Interface) *RelationshipWriter {
	return &RelationshipWriter{
		database: database,
		logger:   logger,
	}
}

// WriteRelationships - Write a Relation to repository
func (r *RelationshipWriter) WriteRelationships(ctx context.Context, tenantID string, collection *database.TupleCollection) (token.EncodedSnapToken, error) {
	var err error

	iterator := collection.CreateTupleIterator()
	if !iterator.HasNext() {
		return token.NewNoopToken().Encode(), nil
	}

	txn := r.database.DB.Txn(true)
	defer txn.Abort()

	for iterator.HasNext() {
		bt := iterator.GetNext()
		t := repositories.RelationTuple{
			ID:              utils.RelationTuplesID.ID(),
			TenantID:        tenantID,
			EntityType:      bt.GetEntity().GetType(),
			EntityID:        bt.GetEntity().GetId(),
			Relation:        bt.GetRelation(),
			SubjectType:     bt.GetSubject().GetType(),
			SubjectID:       bt.GetSubject().GetId(),
			SubjectRelation: bt.GetSubject().GetRelation(),
		}
		if err = txn.Insert(RelationTuplesTable, t); err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	txn.Commit()
	return snapshot.NewToken(time.Now()).Encode(), nil
}

// DeleteRelationships - Delete relationship from repository
func (r *RelationshipWriter) DeleteRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter) (token.EncodedSnapToken, error) {
	var err error
	txn := r.database.DB.Txn(true)
	defer txn.Abort()

	index, args := utils.GetIndexNameAndArgsByFilters(tenantID, filter)
	var it memdb.ResultIterator
	it, err = txn.Get(RelationTuplesTable, index, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	fit := memdb.NewFilterIterator(it, utils.FilterQuery(filter))
	for obj := fit.Next(); obj != nil; obj = fit.Next() {
		t, ok := obj.(repositories.RelationTuple)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		err = txn.Delete(RelationTuplesTable, t)
		if err != nil {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
		}
	}

	txn.Commit()
	return snapshot.NewToken(time.Now()).Encode(), nil
}
