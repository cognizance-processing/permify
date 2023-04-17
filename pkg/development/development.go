package development

import (
	"fmt"

	"permify/internal/config"
	"permify/internal/engines"
	"permify/internal/factories"
	"permify/internal/keys"
	"permify/internal/services"
	"permify/pkg/database"
	"permify/pkg/logger"
)

// Container - Structure for container instance
type Container struct {
	P services.IPermissionService
	R services.IRelationshipService
	S services.ISchemaService
}

// NewContainer - Creates new container instance
func NewContainer() *Container {
	var err error

	var db database.Database
	db, err = factories.DatabaseFactory(config.Database{Engine: database.MEMORY.String()})
	if err != nil {
		fmt.Println(err)
	}

	l := logger.New("debug")

	// Repositories
	relationshipReader := factories.RelationshipReaderFactory(db, l)
	relationshipWriter := factories.RelationshipWriterFactory(db, l)

	schemaReader := factories.SchemaReaderFactory(db, l)
	schemaWriter := factories.SchemaWriterFactory(db, l)

	// engines
	checkEngine := engines.NewCheckEngine(keys.NewNoopCheckEngineKeys(), schemaReader, relationshipReader)
	expandEngine := engines.NewExpandEngine(schemaReader, relationshipReader)
	lookupSchemaEngine := engines.NewLookupSchemaEngine(schemaReader)
	linkedEntityEngine := engines.NewLinkedEntityEngine(schemaReader, relationshipReader)
	lookupEntityEngine := engines.NewLookupEntityEngine(checkEngine, linkedEntityEngine)

	return &Container{
		P: services.NewPermissionService(checkEngine, expandEngine, lookupSchemaEngine, lookupEntityEngine),
		R: services.NewRelationshipService(relationshipReader, relationshipWriter, schemaReader),
		S: services.NewSchemaService(schemaWriter, schemaReader),
	}
}
