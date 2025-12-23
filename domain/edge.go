package domain

import (
	"fmt"
	"strings"
)

// Edge represents an edge connecting two nodes in the graph.
type Edge struct {
	ID          uint64
	Relation    string
	Source      *Node
	Destination *Node
	Properties  map[string]interface{}
	SrcNodeID   uint64
	DestNodeID  uint64
}

// NewEdge creates a new Edge.
func NewEdge(relation string, srcNode *Node, destNode *Node, properties map[string]interface{}) *Edge {
	p := properties
	if p == nil {
		p = make(map[string]interface{})
	}

	return &Edge{
		Relation:    relation,
		Source:      srcNode,
		Destination: destNode,
		Properties:  p,
	}
}

// SetProperty assigns a new property to edge.
func (e *Edge) SetProperty(key string, value interface{}) {
	e.Properties[key] = value
}

// GetProperty retrieves property from edge.
func (e *Edge) GetProperty(key string) interface{} {
	return e.Properties[key]
}

// SourceNodeID returns edge source node ID.
func (e Edge) GetSourceNodeID() uint64 {
	if e.Source != nil {
		return e.Source.ID
	}
	return e.SrcNodeID
}

// DestNodeID returns edge destination node ID.
func (e Edge) GetDestNodeID() uint64 {
	if e.Source != nil {
		return e.Destination.ID
	}
	return e.DestNodeID
}

// String returns a string representation of edge.
func (e Edge) String() string {
	if len(e.Properties) == 0 {
		return "{}"
	}

	p := make([]string, 0, len(e.Properties))
	for k, v := range e.Properties {
		p = append(p, fmt.Sprintf("%s:%v", k, toString(v)))
	}

	return fmt.Sprintf("{%s}", strings.Join(p, ","))
}

// Encode makes Edge satisfy the Stringer interface.
func (e Edge) Encode() string {
	s := []string{"(", e.Source.Alias, ")"}

	s = append(s, "-[")

	if e.Relation != "" {
		s = append(s, ":", e.Relation)
	}

	if len(e.Properties) > 0 {
		p := make([]string, 0, len(e.Properties))
		for k, v := range e.Properties {
			p = append(p, fmt.Sprintf("%s:%v", k, toString(v)))
		}

		s = append(s, "{")
		s = append(s, strings.Join(p, ","))
		s = append(s, "}")
	}

	s = append(s, "]->")
	s = append(s, "(", e.Destination.Alias, ")")

	return strings.Join(s, "")
}
