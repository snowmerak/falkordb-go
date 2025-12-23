package graph

import (
	"errors"
	"fmt"
)

type GraphSchema struct {
	graph         *Graph
	version       int
	labels        []string
	relationships []string
	properties    []string
}

func GraphSchemaNew(graph *Graph) GraphSchema {
	return GraphSchema{
		graph:         graph,
		version:       0,
		labels:        []string{},
		relationships: []string{},
		properties:    []string{},
	}
}

// GraphSchemaWithData seeds schema metadata; primarily used in tests.
func GraphSchemaWithData(labels, relationships, properties []string) GraphSchema {
	return GraphSchema{
		labels:        labels,
		relationships: relationships,
		properties:    properties,
	}
}

func (gs *GraphSchema) clear() {
	gs.labels = []string{}
	gs.relationships = []string{}
	gs.properties = []string{}
}

func (gs *GraphSchema) refresh_labels() error {
	qr, err := gs.graph.CallProcedure("db.labels", nil)
	if err != nil {
		return err
	}

	gs.labels = make([]string, len(qr.results))

	for idx, r := range qr.results {
		val := r.GetByIndex(0)
		s, ok := val.(string)
		if !ok {
			return fmt.Errorf("label name not string: %T", val)
		}
		gs.labels[idx] = s
	}
	return nil
}

func (gs *GraphSchema) refresh_relationships() error {
	qr, err := gs.graph.CallProcedure("db.relationshipTypes", nil)
	if err != nil {
		return err
	}

	gs.relationships = make([]string, len(qr.results))

	for idx, r := range qr.results {
		val := r.GetByIndex(0)
		s, ok := val.(string)
		if !ok {
			return fmt.Errorf("relationship name not string: %T", val)
		}
		gs.relationships[idx] = s
	}
	return nil
}

func (gs *GraphSchema) refresh_properties() error {
	qr, err := gs.graph.CallProcedure("db.propertyKeys", nil)
	if err != nil {
		return err
	}

	gs.properties = make([]string, len(qr.results))

	for idx, r := range qr.results {
		val := r.GetByIndex(0)
		s, ok := val.(string)
		if !ok {
			return fmt.Errorf("property name not string: %T", val)
		}
		gs.properties[idx] = s
	}
	return nil
}

func (gs *GraphSchema) getLabel(lblIdx int) (string, error) {
	if lblIdx >= len(gs.labels) {
		err := gs.refresh_labels()
		if err != nil {
			return "", err
		}
		if lblIdx >= len(gs.labels) {
			return "", errors.New("Unknown label index.")

		}
	}

	return gs.labels[lblIdx], nil
}

func (gs *GraphSchema) getRelation(relIdx int) (string, error) {
	if relIdx >= len(gs.relationships) {
		err := gs.refresh_relationships()
		if err != nil {
			return "", err
		}
		if relIdx >= len(gs.relationships) {
			return "", errors.New("Unknown label index.")
		}
	}

	return gs.relationships[relIdx], nil
}

func (gs *GraphSchema) getProperty(propIdx int) (string, error) {
	if propIdx >= len(gs.properties) {
		err := gs.refresh_properties()
		if err != nil {
			return "", err
		}
		if propIdx >= len(gs.properties) {
			return "", errors.New("Unknown property index.")
		}
	}

	return gs.properties[propIdx], nil
}
