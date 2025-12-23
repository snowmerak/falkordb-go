package domain

import (
	"fmt"
	"strings"
)

type Path struct {
	Nodes []*Node
	Edges []*Edge
}

// NewPath builds a Path from node and edge slices.
func NewPath(nodes []interface{}, edges []interface{}) Path {
	path := Path{Nodes: make([]*Node, len(nodes)), Edges: make([]*Edge, len(edges))}
	for i := 0; i < len(nodes); i++ {
		n, ok := nodes[i].(*Node)
		if !ok {
			return Path{}
		}
		path.Nodes[i] = n
	}
	for i := 0; i < len(edges); i++ {
		e, ok := edges[i].(*Edge)
		if !ok {
			return Path{}
		}
		path.Edges[i] = e
	}

	return path
}

func (p Path) GetNodes() []*Node {
	return p.Nodes
}

func (p Path) GetEdges() []*Edge {
	return p.Edges
}

func (p Path) GetNode(index int) *Node {
	return p.Nodes[index]
}

func (p Path) GetEdge(index int) *Edge {
	return p.Edges[index]
}

func (p Path) FirstNode() *Node {
	return p.GetNode(0)
}

func (p Path) LastNode() *Node {
	return p.GetNode(p.NodesCount() - 1)
}

func (p Path) NodesCount() int {
	return len(p.Nodes)
}

func (p Path) EdgeCount() int {
	return len(p.Edges)
}

func (p Path) String() string {
	s := []string{"<"}
	edgeCount := p.EdgeCount()
	for i := 0; i < edgeCount; i++ {
		node := p.GetNode(i)
		s = append(s, "(", fmt.Sprintf("%v", node.ID), ")")
		edge := p.GetEdge(i)
		if node.ID == edge.SrcNodeID {
			s = append(s, "-[", fmt.Sprintf("%v", edge.ID), "]->")
		} else {
			s = append(s, "<-[", fmt.Sprintf("%v", edge.ID), "]-")
		}
	}
	s = append(s, "(", fmt.Sprintf("%v", p.GetNode(edgeCount).ID), ")")
	s = append(s, ">")

	return strings.Join(s, "")
}
