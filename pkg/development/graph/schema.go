package graph

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/xid"

	"permify/internal/schema"
	base "permify/pkg/pb/base/v1"
	"permify/pkg/tuple"
)

// SchemaToGraph - Convert schema to graph
func SchemaToGraph(schema *base.SchemaDefinition) (g Graph, err error) {
	for _, en := range schema.GetEntityDefinitions() {
		eg, err := EntityToGraph(en)
		if err != nil {
			return Graph{}, err
		}
		g.AddNodes(eg.Nodes())
		g.AddEdges(eg.Edges())
	}
	return
}

// EntityToGraph - Convert entity to graph
func EntityToGraph(entity *base.EntityDefinition) (g Graph, err error) {
	enNode := &Node{
		Type:  "entity",
		ID:    fmt.Sprintf("entity:%s", entity.GetName()),
		Label: entity.GetName(),
	}
	g.AddNode(enNode)

	for _, re := range entity.GetRelations() {
		reNode := &Node{
			Type:  "relation",
			ID:    fmt.Sprintf("entity:%s:permission:%s", entity.GetName(), re.GetName()),
			Label: re.Name,
		}
		g.AddNode(reNode)
		g.AddEdge(enNode, reNode, nil)
	}

	for _, permission := range entity.GetPermissions() {
		acNode := &Node{
			Type:  "permission",
			ID:    fmt.Sprintf("entity:%s:permission:%s", entity.GetName(), permission.GetName()),
			Label: permission.GetName(),
		}
		g.AddNode(acNode)
		g.AddEdge(enNode, acNode, nil)
		ag, err := buildActionGraph(entity, acNode, []*base.Child{permission.GetChild()})
		if err != nil {
			return Graph{}, err
		}
		g.AddNodes(ag.Nodes())
		g.AddEdges(ag.Edges())
	}
	return
}

// buildActionGraph - creates permission graph
func buildActionGraph(entity *base.EntityDefinition, from *Node, children []*base.Child) (g Graph, err error) {
	for _, child := range children {
		switch child.GetType().(type) {
		case *base.Child_Rewrite:
			rw := &Node{
				Type:  "logic",
				ID:    xid.New().String(),
				Label: child.GetRewrite().GetRewriteOperation().String(),
			}

			g.AddNode(rw)
			g.AddEdge(from, rw, nil)
			ag, err := buildActionGraph(entity, rw, child.GetRewrite().GetChildren())
			if err != nil {
				return Graph{}, err
			}
			g.AddNodes(ag.Nodes())
			g.AddEdges(ag.Edges())
		case *base.Child_Leaf:
			leaf := child.GetLeaf()
			switch leaf.GetType().(type) {
			case *base.Leaf_TupleToUserSet:
				re, err := schema.GetRelationByNameInEntityDefinition(entity, leaf.GetTupleToUserSet().GetTupleSet().GetRelation())
				if err != nil {
					return Graph{}, errors.New(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String())
				}
				g.AddEdge(from, &Node{
					Type:  "relation",
					ID:    fmt.Sprintf("entity:%s:permission:%s", GetMainReference(re), leaf.GetTupleToUserSet().GetComputed().GetRelation()),
					Label: leaf.GetTupleToUserSet().GetComputed().GetRelation(),
				}, leaf.GetExclusion())
			case *base.Leaf_ComputedUserSet:
				g.AddEdge(from, &Node{
					Type:  "relation",
					ID:    fmt.Sprintf("entity:%s:permission:%s", entity.GetName(), leaf.GetComputedUserSet().GetRelation()),
					Label: leaf.GetComputedUserSet().GetRelation(),
				}, leaf.GetExclusion())
			default:
				break
			}
		}
	}
	return
}

// GetMainReference -
func GetMainReference(definition *base.RelationDefinition) string {
	for _, ref := range definition.GetRelationReferences() {
		if !strings.Contains(ref.String(), "#") {
			return ref.GetType()
		}
	}
	return tuple.USER
}
