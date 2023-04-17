package engines

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"permify/internal/keys"
	"permify/internal/repositories/mocks"
	"permify/internal/schema"
	"permify/pkg/database"
	base "permify/pkg/pb/base/v1"
	"permify/pkg/token"
	"permify/pkg/tuple"
)

var _ = Describe("check-engine", func() {
	var checkEngine *CheckEngine

	// DRIVE SAMPLE

	driveSchema := `
entity user {}

entity organization {
	relation admin @user
}

entity folder {
	relation org @organization
	relation creator @user
	relation collaborator @user

	permission read = collaborator
	permission update = collaborator
	permission delete = creator or org.admin
}

entity doc {
	relation org @organization
	relation parent @folder
	relation owner @user
	
	permission read = (owner or parent.collaborator) or org.admin
	permission update = owner and org.admin
	permission delete = owner or org.admin
	permission share = update and (owner or parent.update)
}
`

	Context("Drive Sample: Check", func() {
		It("Drive Sample: Case 1", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var doc *base.EntityDefinition
			doc, err = schema.GetEntityByName(sch, "doc")
			Expect(err).ShouldNot(HaveOccurred())

			var folder *base.EntityDefinition
			folder, err = schema.GetEntityByName(sch, "folder")
			Expect(err).ShouldNot(HaveOccurred())

			var organization *base.EntityDefinition
			organization, err = schema.GetEntityByName(sch, "organization")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "t1", "doc", "noop").Return(doc, "noop", nil).Times(2)
			schemaReader.On("ReadSchemaDefinition", "t1", "folder", "noop").Return(folder, "noop", nil).Times(1)
			schemaReader.On("ReadSchemaDefinition", "t1", "organization", "noop").Return(organization, "noop", nil).Times(1)

			// RELATIONSHIPS

			relationshipReader := new(mocks.RelationshipReader)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "doc",
					Ids:  []string{"1"},
				},
				Relation: "owner",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "doc",
					Ids:  []string{"1"},
				},
				Relation: "parent",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "parent",
					Subject: &base.Subject{
						Type:     "folder",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "folder",
					Ids:  []string{"1"},
				},
				Relation: "collaborator",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "3",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "doc",
					Ids:  []string{"1"},
				},
				Relation: "org",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "org",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"1"},
				},
				Relation: "admin",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "1",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			checkEngine = NewCheckEngine(keys.NewNoopCheckEngineKeys(), schemaReader, relationshipReader)

			req := &base.PermissionCheckRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "doc", Id: "1"},
				Subject:    &base.Subject{Type: tuple.USER, Id: "1"},
				Permission: "read",
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "noop",
					Exclusion:     false,
					Depth:         20,
				},
			}

			var response *base.PermissionCheckResponse
			response, err = checkEngine.Run(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(base.PermissionCheckResponse_RESULT_ALLOWED).Should(Equal(response.GetCan()))
		})

		It("Drive Sample: Case 2", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var doc *base.EntityDefinition
			doc, err = schema.GetEntityByName(sch, "doc")
			Expect(err).ShouldNot(HaveOccurred())

			var folder *base.EntityDefinition
			folder, err = schema.GetEntityByName(sch, "folder")
			Expect(err).ShouldNot(HaveOccurred())

			var organization *base.EntityDefinition
			organization, err = schema.GetEntityByName(sch, "organization")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "t1", "doc", "noop").Return(doc, "noop", nil).Times(2)
			schemaReader.On("ReadSchemaDefinition", "t1", "folder", "noop").Return(folder, "noop", nil).Times(1)
			schemaReader.On("ReadSchemaDefinition", "t1", "organization", "noop").Return(organization, "noop", nil).Times(1)

			// RELATIONSHIPS

			relationshipReader := new(mocks.RelationshipReader)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "doc",
					Ids:  []string{"1"},
				},
				Relation: "owner",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "doc",
					Ids:  []string{"1"},
				},
				Relation: "org",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "org",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"1"},
				},
				Relation: "admin",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "1",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			checkEngine = NewCheckEngine(keys.NewNoopCheckEngineKeys(), schemaReader, relationshipReader)

			req := &base.PermissionCheckRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "doc", Id: "1"},
				Subject:    &base.Subject{Type: tuple.USER, Id: "1"},
				Permission: "update",
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "noop",
					Exclusion:     false,
					Depth:         20,
				},
			}

			var response *base.PermissionCheckResponse
			response, err = checkEngine.Run(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(base.PermissionCheckResponse_RESULT_DENIED).Should(Equal(response.GetCan()))
		})

		It("Drive Sample: Case 3", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var doc *base.EntityDefinition
			doc, err = schema.GetEntityByName(sch, "doc")
			Expect(err).ShouldNot(HaveOccurred())

			var folder *base.EntityDefinition
			folder, err = schema.GetEntityByName(sch, "folder")
			Expect(err).ShouldNot(HaveOccurred())

			var organization *base.EntityDefinition
			organization, err = schema.GetEntityByName(sch, "organization")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "t1", "doc", "noop").Return(doc, "noop", nil).Times(2)
			schemaReader.On("ReadSchemaDefinition", "t1", "folder", "noop").Return(folder, "noop", nil).Times(1)
			schemaReader.On("ReadSchemaDefinition", "t1", "organization", "noop").Return(organization, "noop", nil).Times(1)

			// RELATIONSHIPS

			relationshipReader := new(mocks.RelationshipReader)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "doc",
					Ids:  []string{"1"},
				},
				Relation: "owner",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "doc",
					Ids:  []string{"1"},
				},
				Relation: "parent",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "parent",
					Subject: &base.Subject{
						Type:     "folder",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "folder",
					Ids:  []string{"1"},
				},
				Relation: "collaborator",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "7",
						Relation: "",
					},
				},
				{
					Entity: &base.Entity{
						Type: "folder",
						Id:   "1",
					},
					Relation: "collaborator",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "3",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "doc",
					Ids:  []string{"1"},
				},
				Relation: "org",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "org",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"1"},
				},
				Relation: "admin",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "1",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "7",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			checkEngine = NewCheckEngine(keys.NewNoopCheckEngineKeys(), schemaReader, relationshipReader)

			req := &base.PermissionCheckRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "doc", Id: "1"},
				Subject:    &base.Subject{Type: tuple.USER, Id: "1"},
				Permission: "read",
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "noop",
					Exclusion:     false,
					Depth:         20,
				},
			}

			var response *base.PermissionCheckResponse
			response, err = checkEngine.Run(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(base.PermissionCheckResponse_RESULT_DENIED).Should(Equal(response.GetCan()))
		})
	})

	// GITHUB SAMPLE

	githubSchema := `
	entity user {}
	
	entity organization {
		relation admin @user
		relation member @user
	
		action create_repository = admin or member
		action delete = admin
	}
	
	entity repository {
		relation parent @organization
		relation owner @user
	
		action push   = owner
	 action read   = owner and (parent.admin or parent.member)
	 action delete = parent.member and (parent.admin or owner)
	}
	`

	Context("Github Sample: Check", func() {
		It("Github Sample: Case 1", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var repository *base.EntityDefinition
			repository, err = schema.GetEntityByName(sch, "repository")
			Expect(err).ShouldNot(HaveOccurred())

			var organization *base.EntityDefinition
			organization, err = schema.GetEntityByName(sch, "organization")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "t1", "repository", "noop").Return(repository, "noop", nil).Times(2)
			schemaReader.On("ReadSchemaDefinition", "t1", "organization", "noop").Return(organization, "noop", nil).Times(2)

			// RELATIONSHIPS

			relationshipReader := new(mocks.RelationshipReader)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "repository",
					Ids:  []string{"1"},
				},
				Relation: "owner",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "repository",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			checkEngine = NewCheckEngine(keys.NewNoopCheckEngineKeys(), schemaReader, relationshipReader)

			req := &base.PermissionCheckRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "repository", Id: "1"},
				Subject:    &base.Subject{Type: tuple.USER, Id: "1"},
				Permission: "push",
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "noop",
					Exclusion:     false,
					Depth:         20,
				},
			}

			var response *base.PermissionCheckResponse
			response, err = checkEngine.Run(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(base.PermissionCheckResponse_RESULT_DENIED).Should(Equal(response.GetCan()))
		})

		It("Github Sample: Case 2", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var repository *base.EntityDefinition
			repository, err = schema.GetEntityByName(sch, "repository")
			Expect(err).ShouldNot(HaveOccurred())

			var organization *base.EntityDefinition
			organization, err = schema.GetEntityByName(sch, "organization")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "t1", "repository", "noop").Return(repository, "noop", nil).Times(2)
			schemaReader.On("ReadSchemaDefinition", "t1", "organization", "noop").Return(organization, "noop", nil).Times(2)

			// RELATIONSHIPS

			relationshipReader := new(mocks.RelationshipReader)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "repository",
					Ids:  []string{"1"},
				},
				Relation: "owner",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "repository",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "2",
						Relation: "admin",
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"2"},
				},
				Relation: "admin",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "2",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "3",
						Relation: "member",
					},
				},
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "2",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "3",
						Relation: "",
					},
				},
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "2",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "8",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"3"},
				},
				Relation: "member",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "3",
					},
					Relation: "member",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			checkEngine = NewCheckEngine(keys.NewNoopCheckEngineKeys(), schemaReader, relationshipReader)

			req := &base.PermissionCheckRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "repository", Id: "1"},
				Subject:    &base.Subject{Type: tuple.USER, Id: "1"},
				Permission: "push",
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "noop",
					Exclusion:     false,
					Depth:         20,
				},
			}

			var response *base.PermissionCheckResponse
			response, err = checkEngine.Run(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(base.PermissionCheckResponse_RESULT_ALLOWED).Should(Equal(response.GetCan()))
		})

		It("Github Sample: Case 3", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var repository *base.EntityDefinition
			repository, err = schema.GetEntityByName(sch, "repository")
			Expect(err).ShouldNot(HaveOccurred())

			var organization *base.EntityDefinition
			organization, err = schema.GetEntityByName(sch, "organization")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "t1", "repository", "noop").Return(repository, "noop", nil).Times(2)
			schemaReader.On("ReadSchemaDefinition", "t1", "organization", "noop").Return(organization, "noop", nil).Times(2)

			// RELATIONSHIPS

			relationshipReader := new(mocks.RelationshipReader)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "repository",
					Ids:  []string{"1"},
				},
				Relation: "parent",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "repository",
						Id:   "1",
					},
					Relation: "parent",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "8",
						Relation: tuple.ELLIPSIS,
					},
				},
			}...), nil).Times(2)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"8"},
				},
				Relation: "member",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "8",
					},
					Relation: "member",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"8"},
				},
				Relation: "admin",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "8",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "repository",
					Ids:  []string{"1"},
				},
				Relation: "owner",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "repository",
						Id:   "1",
					},
					Relation: "owner",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "7",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			checkEngine = NewCheckEngine(keys.NewNoopCheckEngineKeys(), schemaReader, relationshipReader)

			req := &base.PermissionCheckRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "repository", Id: "1"},
				Subject:    &base.Subject{Type: tuple.USER, Id: "1"},
				Permission: "delete",
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "noop",
					Exclusion:     false,
					Depth:         20,
				},
			}

			var response *base.PermissionCheckResponse
			response, err = checkEngine.Run(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(base.PermissionCheckResponse_RESULT_DENIED).Should(Equal(response.GetCan()))
		})
	})

	// EXCLUSION SAMPLE

	exclusionSchema := `
	entity user {}
        
	entity organization {
    	relation member @user
	}

	entity parent {
    	relation member @user
	}

	entity repo {

    	relation org @organization
    	relation parent @parent
    
    	permission push   = org.member and not parent.member

	} 
	`

	Context("Exclusion Sample: Check", func() {
		It("Exclusion Sample: Case 1", func() {
			var err error

			// SCHEMA

			schemaReader := new(mocks.SchemaReader)

			var sch *base.SchemaDefinition
			sch, err = schema.NewSchemaFromStringDefinitions(true, exclusionSchema)
			Expect(err).ShouldNot(HaveOccurred())

			var repo *base.EntityDefinition
			repo, err = schema.GetEntityByName(sch, "repo")
			Expect(err).ShouldNot(HaveOccurred())

			var parent *base.EntityDefinition
			parent, err = schema.GetEntityByName(sch, "parent")
			Expect(err).ShouldNot(HaveOccurred())

			var organization *base.EntityDefinition
			organization, err = schema.GetEntityByName(sch, "organization")
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader.On("ReadSchemaDefinition", "t1", "repo", "noop").Return(repo, "noop", nil).Times(1)
			schemaReader.On("ReadSchemaDefinition", "t1", "parent", "noop").Return(parent, "noop", nil).Times(1)
			schemaReader.On("ReadSchemaDefinition", "t1", "organization", "noop").Return(organization, "noop", nil).Times(1)

			// RELATIONSHIPS

			relationshipReader := new(mocks.RelationshipReader)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"1"},
				},
				Relation: "member",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "1",
					},
					Relation: "member",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "1",
					},
					Relation: "member",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "2",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "parent",
					Ids:  []string{"1"},
				},
				Relation: "member",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "parent",
						Id:   "1",
					},
					Relation: "member",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "1",
						Relation: "",
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "repo",
					Ids:  []string{"1"},
				},
				Relation: "org",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "repo",
						Id:   "1",
					},
					Relation: "org",
					Subject: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}...), nil).Times(1)

			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "repo",
					Ids:  []string{"1"},
				},
				Relation: "parent",
			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "repo",
						Id:   "1",
					},
					Relation: "parent",
					Subject: &base.Subject{
						Type:     "parent",
						Id:       "1",
						Relation: tuple.ELLIPSIS,
					},
				},
			}...), nil).Times(1)

			checkEngine = NewCheckEngine(keys.NewNoopCheckEngineKeys(), schemaReader, relationshipReader)

			req := &base.PermissionCheckRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "repo", Id: "1"},
				Subject:    &base.Subject{Type: tuple.USER, Id: "2"},
				Permission: "push",
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "noop",
					Exclusion:     false,
					Depth:         20,
				},
			}

			var response *base.PermissionCheckResponse
			response, err = checkEngine.Run(context.Background(), req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(base.PermissionCheckResponse_RESULT_ALLOWED).Should(Equal(response.GetCan()))
		})
	})
})
