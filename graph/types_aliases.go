package graph

import "github.com/snowmerak/falkordb-go/domain"

// Re-export domain types for graph package consumers.
type (
	Node   = domain.Node
	Edge   = domain.Edge
	Path   = domain.Path
	Record = domain.Record
)

func NodeNew(labels []string, alias string, properties map[string]interface{}) *Node {
	return domain.NodeNew(labels, alias, properties)
}

func EdgeNew(relation string, srcNode *Node, destNode *Node, properties map[string]interface{}) *Edge {
	return domain.EdgeNew(relation, srcNode, destNode, properties)
}

func PathNew(nodes []interface{}, edges []interface{}) Path {
	return domain.PathNew(nodes, edges)
}

func recordNew(values []interface{}, keys []string) *Record {
	return domain.NewRecord(values, keys)
}
