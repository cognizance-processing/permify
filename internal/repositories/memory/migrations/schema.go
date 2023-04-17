package migrations

import (
	"github.com/hashicorp/go-memdb"

	"permify/internal/repositories/memory"
)

// Schema - Database schema for memory db
var Schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		memory.SchemaDefinitionsTable: {
			Name: memory.SchemaDefinitionsTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "Version"},
						},
					},
				},
				"version": {
					Name:   "version",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "Version"},
						},
					},
				},
				"tenant": {
					Name:   "tenant",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
						},
					},
				},
			},
		},
		memory.RelationTuplesTable: {
			Name: memory.RelationTuplesTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "EntityID"},
							&memdb.StringFieldIndex{Field: "Relation"},
							&memdb.StringFieldIndex{Field: "SubjectType"},
							&memdb.StringFieldIndex{Field: "SubjectID"},
							&memdb.StringFieldIndex{Field: "SubjectRelation"},
						},
						AllowMissing: true,
					},
				},
				"entity-index": {
					Name:   "entity-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "EntityID"},
							&memdb.StringFieldIndex{Field: "Relation"},
						},
					},
				},
				"relation-index": {
					Name:   "relation-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "Relation"},
							&memdb.StringFieldIndex{Field: "SubjectType"},
						},
					},
				},
				"entity-type-index": {
					Name:   "entity-type-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
						},
					},
				},
				"entity-type-and-relation-index": {
					Name:   "entity-type-and-relation-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "Relation"},
						},
					},
				},
			},
		},
		memory.TenantsTable: {
			Name: memory.TenantsTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "ID"},
						},
					},
				},
			},
		},
	},
}
