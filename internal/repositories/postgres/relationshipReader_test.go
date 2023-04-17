package postgres

import (
	"context"
	"database/sql"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgtype"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"permify/internal/repositories/postgres/snapshot"
	"permify/internal/repositories/postgres/types"
	"permify/pkg/database"
	"permify/pkg/database/postgres"
	"permify/pkg/logger"
	base "permify/pkg/pb/base/v1"
	"permify/pkg/tuple"
)

var _ = Describe("RelationshipReader", func() {
	var relationshipReader *RelationshipReader
	var mock sqlmock.Sqlmock

	BeforeEach(func() {
		l := logger.New("debug")

		var db *sql.DB
		var err error

		db, mock, err = sqlmock.New()
		Expect(err).ShouldNot(HaveOccurred())

		pg := &postgres.Postgres{
			DB:      db,
			Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		}

		relationshipReader = NewRelationshipReader(pg, l)
	})

	AfterEach(func() {
		err := mock.ExpectationsWereMet()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("QueryRelationships", func() {
		columns := []string{"entity_type", "entity_id", "relation", "subject_type", "subject_id", "subject_relation"}

		It("should be same queries", func() {
			rows := sqlmock.NewRows(columns).
				AddRow("organization", "abc", "admin", "user", "jack", "").
				AddRow("organization", "abc", "admin", "user", "john", "")

			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT entity_type, entity_id, relation, subject_type, subject_id, subject_relation
			 FROM relation_tuples WHERE tenant_id = $1 AND entity_id IN ($2) AND entity_type = $3 AND relation = $4 AND (pg_visible_in_snapshot(created_tx_id, 
				(select snapshot from transactions where id = '4'::xid8)) = true OR created_tx_id = '4'::xid8) AND ((pg_visible_in_snapshot(expired_tx_id, 
					(select snapshot from transactions where id = '4'::xid8)) = false OR expired_tx_id = '0'::xid8) AND expired_tx_id <> '4'::xid8)`)).
				WithArgs("noop", "abc", "organization", "admin").
				WillReturnRows(rows)
			mock.ExpectCommit()

			value, err := relationshipReader.QueryRelationships(context.Background(), "noop", &base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"abc"},
				},
				Relation: "admin",
			}, snapshot.NewToken(types.XID8{Uint: 4, Status: pgtype.Present}).Encode().String())

			Expect(err).ShouldNot(HaveOccurred())
			Expect(value).Should(Equal(database.NewTupleIterator([]*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "abc",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "jack",
						Relation: "",
					},
				},
				{
					Entity: &base.Entity{
						Type: "organization",
						Id:   "abc",
					},
					Relation: "admin",
					Subject: &base.Subject{
						Type:     tuple.USER,
						Id:       "john",
						Relation: "",
					},
				},
			}...)))
		})
	})
})
