package database

import (
	base "permify/pkg/pb/base/v1"
)

// TupleCollection -Tuple collection.
type TupleCollection struct {
	tuples []*base.Tuple
}

// NewTupleCollection - Create new tuple collection.
func NewTupleCollection(tuples ...*base.Tuple) *TupleCollection {
	if len(tuples) == 0 {
		return &TupleCollection{}
	}
	return &TupleCollection{
		tuples: tuples,
	}
}

// CreateTupleIterator - Create tuple iterator according to collection.
func (t *TupleCollection) CreateTupleIterator() *TupleIterator {
	return &TupleIterator{
		tuples: t.tuples,
	}
}

// GetTuples - Get tuples
func (t *TupleCollection) GetTuples() []*base.Tuple {
	return t.tuples
}

// Add - New subject to collection.
func (t *TupleCollection) Add(tuple *base.Tuple) {
	t.tuples = append(t.tuples, tuple)
}

// ToSubjectCollection - Converts new subject collection from given tuple collection
func (t *TupleCollection) ToSubjectCollection() *SubjectCollection {
	subjects := make([]*base.Subject, len(t.tuples))
	for index, tuple := range t.tuples {
		subjects[index] = tuple.GetSubject()
	}
	return NewSubjectCollection(subjects...)
}

// SUBJECT

// SubjectCollection - Subject collection.
type SubjectCollection struct {
	subjects []*base.Subject
}

// NewSubjectCollection - Create new subject collection.
func NewSubjectCollection(subjects ...*base.Subject) *SubjectCollection {
	if len(subjects) == 0 {
		return &SubjectCollection{}
	}
	return &SubjectCollection{
		subjects: subjects,
	}
}

// CreateSubjectIterator - Create subject iterator according to collection.
func (s *SubjectCollection) CreateSubjectIterator() *SubjectIterator {
	return &SubjectIterator{
		subjects: s.subjects,
	}
}

// GetSubjects - Get subject collection
func (s *SubjectCollection) GetSubjects() []*base.Subject {
	return s.subjects
}

// Add - New subject to collection.
func (s *SubjectCollection) Add(subject *base.Subject) {
	s.subjects = append(s.subjects, subject)
}

// ENTITY

// EntityCollection - Entity collection.
type EntityCollection struct {
	entities []*base.Entity
}

// NewEntityCollection - Create new subject collection.
func NewEntityCollection(entities ...*base.Entity) *EntityCollection {
	if len(entities) == 0 {
		return &EntityCollection{}
	}
	return &EntityCollection{
		entities: entities,
	}
}

// CreateEntityIterator  - Create entity iterator according to collection.
func (e *EntityCollection) CreateEntityIterator() *EntityIterator {
	return &EntityIterator{
		entities: e.entities,
	}
}

// GetEntities - Get entities
func (e *EntityCollection) GetEntities() []*base.Entity {
	return e.entities
}

// Add - New subject to collection.
func (e *EntityCollection) Add(entity *base.Entity) {
	e.entities = append(e.entities, entity)
}
