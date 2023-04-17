package servers

import (
	"errors"

	"google.golang.org/grpc/status"

	otelCodes "go.opentelemetry.io/otel/codes"
	"golang.org/x/net/context"

	"permify/internal/services"
	"permify/pkg/logger"
	v1 "permify/pkg/pb/base/v1"
	"permify/pkg/tuple"
)

// RelationshipServer - Structure for Relationship Server
type RelationshipServer struct {
	v1.UnimplementedRelationshipServer

	relationshipService services.IRelationshipService
	logger              logger.Interface
}

// NewRelationshipServer - Creates new Relationship Server
func NewRelationshipServer(r services.IRelationshipService, l logger.Interface) *RelationshipServer {
	return &RelationshipServer{
		relationshipService: r,
		logger:              l,
	}
}

// Read - Allows directly querying the stored engines data to display and filter stored relational tuples
func (r *RelationshipServer) Read(ctx context.Context, request *v1.RelationshipReadRequest) (*v1.RelationshipReadResponse, error) {
	ctx, span := tracer.Start(ctx, "relationships.read")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	collection, ct, err := r.relationshipService.ReadRelationships(ctx, request.GetTenantId(), request.GetFilter(), request.GetMetadata().GetSnapToken(), request.GetPageSize(), request.GetContinuousToken())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.RelationshipReadResponse{
		Tuples:          collection.GetTuples(),
		ContinuousToken: ct.String(),
	}, nil
}

// Write - Write relation tuples to writeDB
func (r *RelationshipServer) Write(ctx context.Context, request *v1.RelationshipWriteRequest) (*v1.RelationshipWriteResponse, error) {
	ctx, span := tracer.Start(ctx, "relationships.write")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	for _, tup := range request.GetTuples() {
		if tuple.IsSubjectUser(tup.GetSubject()) {
			if tup.GetSubject().GetRelation() != "" {
				return nil, errors.New(v1.ErrorCode_ERROR_CODE_SUBJECT_RELATION_MUST_BE_EMPTY.String())
			}
		}
	}

	snap, err := r.relationshipService.WriteRelationships(ctx, request.GetTenantId(), request.GetTuples(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.RelationshipWriteResponse{
		SnapToken: snap.String(),
	}, nil
}

// Delete - Delete relation tuples to writeDB
func (r *RelationshipServer) Delete(ctx context.Context, request *v1.RelationshipDeleteRequest) (*v1.RelationshipDeleteResponse, error) {
	ctx, span := tracer.Start(ctx, "relationships.delete")
	defer span.End()

	v := request.Validate()
	if v != nil {
		return nil, v
	}

	snap, err := r.relationshipService.DeleteRelationships(ctx, request.GetTenantId(), request.GetFilter())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		r.logger.Error(err.Error())
		return nil, status.Error(GetStatus(err), err.Error())
	}

	return &v1.RelationshipDeleteResponse{
		SnapToken: snap.String(),
	}, nil
}
