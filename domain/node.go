package domain

import (
	"fmt"
	"strings"
)

// Node represents a node within a graph.
type Node struct {
	ID         uint64
	Labels     []string
	Alias      string
	Properties map[string]interface{}
}

// NodeNew creates a new Node.
func NodeNew(labels []string, alias string, properties map[string]interface{}) *Node {
	p := properties
	if p == nil {
		p = make(map[string]interface{})
	}

	return &Node{
		Labels:     labels,
		Alias:      alias,
		Properties: p,
	}
}

// SetProperty assigns a new property to node.
func (n *Node) SetProperty(key string, value interface{}) {
	n.Properties[key] = value
}

// GetProperty retrieves property from node.
func (n Node) GetProperty(key string) interface{} {
	return n.Properties[key]
}

// String returns a string representation of a node.
func (n Node) String() string {
	if len(n.Properties) == 0 {
		return "{}"
	}

	p := make([]string, 0, len(n.Properties))
	for k, v := range n.Properties {
		p = append(p, fmt.Sprintf("%s:%v", k, toString(v)))
	}

	return fmt.Sprintf("{%s}", strings.Join(p, ","))
}

// Encode makes Node satisfy the Stringer interface.
func (n Node) Encode() string {
	s := []string{"("}

	if n.Alias != "" {
		s = append(s, n.Alias)
	}

	for _, label := range n.Labels {
		s = append(s, ":", label)
	}

	if len(n.Properties) > 0 {
		p := make([]string, 0, len(n.Properties))
		for k, v := range n.Properties {
			p = append(p, fmt.Sprintf("%s:%v", k, toString(v)))
		}

		s = append(s, "{")
		s = append(s, strings.Join(p, ","))
		s = append(s, "}")
	}

	s = append(s, ")")
	return strings.Join(s, "")
}
