package graph

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

const (
	LABELS_ADDED            string = "Labels added"
	NODES_CREATED           string = "Nodes created"
	NODES_DELETED           string = "Nodes deleted"
	RELATIONSHIPS_DELETED   string = "Relationships deleted"
	PROPERTIES_SET          string = "Properties set"
	RELATIONSHIPS_CREATED   string = "Relationships created"
	INDICES_CREATED         string = "Indices created"
	INDICES_DELETED         string = "Indices deleted"
	INTERNAL_EXECUTION_TIME string = "Query internal execution time"
	CACHED_EXECUTION        string = "Cached execution"
)

type ResultSetColumnTypes int

const (
	COLUMN_UNKNOWN ResultSetColumnTypes = iota
	COLUMN_SCALAR
	COLUMN_NODE
	COLUMN_RELATION
)

type ResultSetScalarTypes int

const (
	VALUE_UNKNOWN ResultSetScalarTypes = iota
	VALUE_NULL
	VALUE_STRING
	VALUE_INTEGER
	VALUE_BOOLEAN
	VALUE_DOUBLE
	VALUE_ARRAY
	VALUE_EDGE
	VALUE_NODE
	VALUE_PATH
	VALUE_MAP
	VALUE_POINT
	VALUE_VECTORF32
)

type QueryResultHeader struct {
	column_names []string
	column_types []ResultSetColumnTypes
}

// QueryResult represents the results of a query.
type QueryResult struct {
	graph            *Graph
	header           QueryResultHeader
	results          []*Record
	statistics       map[string]float64
	currentRecordIdx int
}

// Graph returns the graph associated with this result set.
func (qr *QueryResult) Graph() *Graph { return qr.graph }

// Header returns the parsed result header metadata.
func (qr *QueryResult) Header() QueryResultHeader { return qr.header }

// Results returns the raw records slice.
func (qr *QueryResult) Results() []*Record { return qr.results }

// Statistics returns query execution statistics.
func (qr *QueryResult) Statistics() map[string]float64 { return qr.statistics }

// CurrentRecordIndex returns the current cursor position, or -1 if iteration has not started.
func (qr *QueryResult) CurrentRecordIndex() int { return qr.currentRecordIdx }

func QueryResultNew(g *Graph, response interface{}) (*QueryResult, error) {
	qr := &QueryResult{
		results:    nil,
		statistics: nil,
		header: QueryResultHeader{
			column_names: make([]string, 0),
			column_types: make([]ResultSetColumnTypes, 0),
		},
		graph:            g,
		currentRecordIdx: -1,
	}

	r, ok := response.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type %T", response)
	}

	if len(r) == 0 {
		return nil, errors.New("empty response payload")
	}

	if len(r) == 1 {
		stats, ok := r[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("statistics payload is not array: %T", r[0])
		}
		if err := qr.parseStatistics(stats); err != nil {
			return nil, err
		}
		return qr, nil
	}

	if len(r) < 3 {
		return nil, fmt.Errorf("unexpected response length %d", len(r))
	}

	if err := qr.parseResults(r); err != nil {
		return nil, err
	}
	if err := qr.parseStatistics(r[2]); err != nil {
		return nil, err
	}

	return qr, nil
}

func (qr *QueryResult) Empty() bool {
	return len(qr.results) == 0
}

func (qr *QueryResult) parseResults(raw_result_set []interface{}) error {
	if len(raw_result_set) < 2 {
		return errors.New("result set missing header or records")
	}

	header := raw_result_set[0]
	if err := qr.parseHeader(header); err != nil {
		return err
	}

	if err := qr.parseRecords(raw_result_set); err != nil {
		return err
	}

	return nil
}

func (qr *QueryResult) parseStatistics(raw_statistics interface{}) error {
	statistics, ok := raw_statistics.([]interface{})
	if !ok {
		return fmt.Errorf("statistics payload is not array: %T", raw_statistics)
	}
	qr.statistics = make(map[string]float64)

	for _, rs := range statistics {
		rsStr, ok := rs.(string)
		if !ok {
			return fmt.Errorf("statistic entry not string: %T", rs)
		}
		parts := strings.SplitN(rsStr, ": ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid statistic format: %s", rsStr)
		}
		numPart := strings.Split(parts[1], " ")
		if len(numPart) == 0 || numPart[0] == "" {
			return fmt.Errorf("missing statistic value in: %s", rsStr)
		}
		f, err := strconv.ParseFloat(numPart[0], 64)
		if err != nil {
			return fmt.Errorf("invalid statistic value %q: %w", numPart[0], err)
		}
		qr.statistics[parts[0]] = f
	}

	return nil
}

func (qr *QueryResult) parseHeader(raw_header interface{}) error {
	header, ok := raw_header.([]interface{})
	if !ok {
		return fmt.Errorf("header payload is not array: %T", raw_header)
	}

	for _, col := range header {
		c, ok := col.([]interface{})
		if !ok || len(c) < 2 {
			return errors.New("invalid header column format")
		}
		ct, ok := c[0].(int64)
		if !ok {
			return errors.New("header column type is not int64")
		}
		cn, ok := c[1].(string)
		if !ok {
			return errors.New("header column name is not string")
		}

		qr.header.column_types = append(qr.header.column_types, ResultSetColumnTypes(ct))
		qr.header.column_names = append(qr.header.column_names, cn)
	}

	if len(qr.header.column_types) != len(qr.header.column_names) {
		return errors.New("header column metadata length mismatch")
	}

	return nil
}

func (qr *QueryResult) parseRecords(raw_result_set []interface{}) error {
	records, ok := raw_result_set[1].([]interface{})
	if !ok {
		return fmt.Errorf("records payload is not array: %T", raw_result_set[1])
	}

	if len(qr.header.column_types) != len(qr.header.column_names) {
		return errors.New("header metadata not initialized")
	}

	qr.results = make([]*Record, len(records))

	for i, r := range records {
		cells, ok := r.([]interface{})
		if !ok {
			return fmt.Errorf("record %d is not array", i)
		}
		if len(cells) != len(qr.header.column_types) {
			return fmt.Errorf("record %d column count mismatch: got %d want %d", i, len(cells), len(qr.header.column_types))
		}

		values := make([]interface{}, len(cells))

		for idx, c := range cells {
			t := qr.header.column_types[idx]
			switch t {
			case COLUMN_SCALAR:
				cval, ok := c.([]interface{})
				if !ok {
					return fmt.Errorf("record %d column %d not scalar payload", i, idx)
				}
				s, err := qr.parseScalar(cval)
				if err != nil {
					return err
				}
				values[idx] = s
			case COLUMN_NODE:
				v, err := qr.parseNode(c)
				if err != nil {
					return err
				}
				values[idx] = v
			case COLUMN_RELATION:
				v, err := qr.parseEdge(c)
				if err != nil {
					return err
				}
				values[idx] = v
			default:
				return errors.New("unknown column type")
			}
		}
		qr.results[i] = recordNew(values, qr.header.column_names)
	}
	return nil
}

func (qr *QueryResult) parseProperties(props []interface{}) (map[string]interface{}, error) {
	// [[name, value type, value] X N]
	properties := make(map[string]interface{})
	for _, prop := range props {
		p, ok := prop.([]interface{})
		if !ok || len(p) < 3 {
			return nil, errors.New("invalid property format")
		}
		idx, ok := p[0].(int64)
		if !ok {
			return nil, errors.New("property index not int64")
		}
		prop_name, err := qr.graph.schema.getProperty(int(idx))
		if err != nil {
			return nil, err
		}
		prop_value, err := qr.parseScalar(p[1:])
		if err != nil {
			return nil, err
		}
		properties[prop_name] = prop_value
	}

	return properties, nil
}

func (qr *QueryResult) parseNode(cell interface{}) (*Node, error) {
	// Node ID (integer),
	// [label string offset (integer)],
	// [[name, value type, value] X N]

	c, ok := cell.([]interface{})
	if !ok || len(c) < 3 {
		return nil, errors.New("invalid node payload")
	}
	id, ok := c[0].(int64)
	if !ok {
		return nil, errors.New("node id not int64")
	}
	labelIds, ok := c[1].([]interface{})
	if !ok {
		return nil, errors.New("node labels not array")
	}
	labels := make([]string, len(labelIds))
	for i := 0; i < len(labelIds); i++ {
		lid, ok := labelIds[i].(int64)
		if !ok {
			return nil, errors.New("label id not int64")
		}
		label, err := qr.graph.schema.getLabel(int(lid))
		if err != nil {
			return nil, err
		}
		labels[i] = label
	}

	rawProps, ok := c[2].([]interface{})
	if !ok {
		return nil, errors.New("node properties not array")
	}
	properties, err := qr.parseProperties(rawProps)
	if err != nil {
		return nil, err
	}

	n := NodeNew(labels, "", properties)
	n.ID = uint64(id)
	return n, nil
}

func (qr *QueryResult) parseEdge(cell interface{}) (*Edge, error) {
	// Edge ID (integer),
	// reltype string offset (integer),
	// src node ID offset (integer),
	// dest node ID offset (integer),
	// [[name, value, value type] X N]

	c, ok := cell.([]interface{})
	if !ok || len(c) < 5 {
		return nil, errors.New("invalid edge payload")
	}
	id, ok := c[0].(int64)
	if !ok {
		return nil, errors.New("edge id not int64")
	}
	r, ok := c[1].(int64)
	if !ok {
		return nil, errors.New("edge relation id not int64")
	}
	relation, err := qr.graph.schema.getRelation(int(r))
	if err != nil {
		return nil, err
	}

	src_node_id, ok := c[2].(int64)
	if !ok {
		return nil, errors.New("edge src id not int64")
	}
	dest_node_id, ok := c[3].(int64)
	if !ok {
		return nil, errors.New("edge dest id not int64")
	}
	rawProps, ok := c[4].([]interface{})
	if !ok {
		return nil, errors.New("edge properties not array")
	}
	properties, err := qr.parseProperties(rawProps)
	if err != nil {
		return nil, err
	}
	e := EdgeNew(relation, nil, nil, properties)

	e.ID = uint64(id)
	e.SrcNodeID = uint64(src_node_id)
	e.DestNodeID = uint64(dest_node_id)
	return e, nil
}

func (qr *QueryResult) parseArray(cell interface{}) ([]interface{}, error) {
	array, ok := cell.([]interface{})
	if !ok {
		return nil, errors.New("array payload is not array")
	}
	var arrayLength = len(array)
	for i := 0; i < arrayLength; i++ {
		inner, ok := array[i].([]interface{})
		if !ok {
			return nil, fmt.Errorf("array element %d not scalar payload", i)
		}
		s, err := qr.parseScalar(inner)
		if err != nil {
			return nil, err
		}
		array[i] = s
	}
	return array, nil
}

func (qr *QueryResult) parsePath(cell interface{}) (Path, error) {
	arrays, ok := cell.([]interface{})
	if !ok || len(arrays) < 2 {
		return Path{}, errors.New("path payload invalid")
	}
	nodesRaw, ok := arrays[0].([]interface{})
	if !ok {
		return Path{}, errors.New("path nodes payload invalid")
	}
	edgesRaw, ok := arrays[1].([]interface{})
	if !ok {
		return Path{}, errors.New("path edges payload invalid")
	}
	nodesVal, err := qr.parseScalar(nodesRaw)
	if err != nil {
		return Path{}, err
	}
	edgesVal, err := qr.parseScalar(edgesRaw)
	if err != nil {
		return Path{}, err
	}
	nodesSlice, ok := nodesVal.([]interface{})
	if !ok {
		return Path{}, errors.New("parsed path nodes not array")
	}
	edgesSlice, ok := edgesVal.([]interface{})
	if !ok {
		return Path{}, errors.New("parsed path edges not array")
	}

	path := Path{Nodes: make([]*Node, len(nodesSlice)), Edges: make([]*Edge, len(edgesSlice))}
	for i := range nodesSlice {
		n, ok := nodesSlice[i].(*Node)
		if !ok {
			return Path{}, errors.New("path node element not *Node")
		}
		path.Nodes[i] = n
	}
	for i := range edgesSlice {
		e, ok := edgesSlice[i].(*Edge)
		if !ok {
			return Path{}, errors.New("path edge element not *Edge")
		}
		path.Edges[i] = e
	}

	return path, nil
}

func (qr *QueryResult) parseMap(cell interface{}) (map[string]interface{}, error) {
	raw_map, ok := cell.([]interface{})
	if !ok {
		return nil, errors.New("map payload not array")
	}
	var mapLength = len(raw_map)
	var parsed_map = make(map[string]interface{})

	if mapLength%2 != 0 {
		return nil, errors.New("map payload length is not even")
	}

	for i := 0; i < mapLength; i += 2 {
		key, ok := raw_map[i].(string)
		if !ok {
			return nil, errors.New("map key not string")
		}
		valRaw, ok := raw_map[i+1].([]interface{})
		if !ok {
			return nil, errors.New("map value payload not scalar array")
		}
		s, err := qr.parseScalar(valRaw)
		if err != nil {
			return nil, err
		}
		parsed_map[key] = s
	}

	return parsed_map, nil
}

func (qr *QueryResult) parsePoint(cell interface{}) (map[string]interface{}, error) {
	parsed_point := make(map[string]interface{})
	array, ok := cell.([]interface{})
	if !ok || len(array) < 2 {
		return nil, errors.New("point payload invalid")
	}
	latStr, ok := array[0].(string)
	if !ok {
		return nil, errors.New("point latitude not string")
	}
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %w", err)
	}
	parsed_point["latitude"] = lat
	lonStr, ok := array[1].(string)
	if !ok {
		return nil, errors.New("point longitude not string")
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %w", err)
	}
	parsed_point["longitude"] = lon
	return parsed_point, nil
}

func (qr *QueryResult) parseVectorF32(cell interface{}) ([]float32, error) {
	array, ok := cell.([]interface{})
	if !ok {
		return nil, errors.New("vector payload not array")
	}
	var arrayLength = len(array)
	var res = make([]float32, arrayLength)
	for i := 0; i < arrayLength; i++ {
		f, ok := array[i].(float64)
		if !ok {
			return nil, errors.New("vector element not float64")
		}
		res[i] = float32(f)
	}
	return res, nil
}

func (qr *QueryResult) parseScalar(cell []interface{}) (interface{}, error) {
	if len(cell) < 2 {
		return nil, errors.New("scalar cell too short")
	}
	t, ok := cell[0].(int64)
	if !ok {
		return nil, errors.New("scalar type not int64")
	}
	v := cell[1]
	switch ResultSetScalarTypes(t) {
	case VALUE_NULL:
		return nil, nil

	case VALUE_STRING:
		s, ok := v.(string)
		if !ok {
			return nil, errors.New("string scalar not string")
		}
		return s, nil

	case VALUE_INTEGER:
		i, ok := v.(int64)
		if !ok {
			return nil, errors.New("integer scalar not int64")
		}
		return i, nil

	case VALUE_BOOLEAN:
		s, ok := v.(string)
		if !ok {
			return nil, errors.New("boolean scalar not string")
		}
		return s == "true", nil

	case VALUE_DOUBLE:
		s, ok := v.(string)
		if !ok {
			return nil, errors.New("double scalar not string")
		}
		return strconv.ParseFloat(s, 64)

	case VALUE_ARRAY:
		return qr.parseArray(v)

	case VALUE_EDGE:
		return qr.parseEdge(v)

	case VALUE_NODE:
		return qr.parseNode(v)

	case VALUE_PATH:
		return qr.parsePath(v)

	case VALUE_MAP:
		return qr.parseMap(v)

	case VALUE_POINT:
		return qr.parsePoint(v)

	case VALUE_VECTORF32:
		return qr.parseVectorF32(v)

	case VALUE_UNKNOWN:
		return nil, errors.New("unknown scalar type")
	}

	return nil, errors.New("unknown scalar type")
}

func (qr *QueryResult) getStat(stat string) float64 {
	if val, ok := qr.statistics[stat]; ok {
		return val
	} else {
		return 0.0
	}
}

// Next returns true only if there is a record to be processed.
func (qr *QueryResult) Next() bool {
	if qr.Empty() {
		return false
	}
	if qr.currentRecordIdx < len(qr.results)-1 {
		qr.currentRecordIdx++
		return true
	} else {
		return false
	}
}

// Record returns the current record.
func (qr *QueryResult) Record() *Record {
	if qr.currentRecordIdx >= 0 && qr.currentRecordIdx < len(qr.results) {
		return qr.results[qr.currentRecordIdx]
	} else {
		return nil
	}
}

// PrettyPrint prints the QueryResult to stdout, pretty-like.
func (qr *QueryResult) PrettyPrint() {
	if qr.Empty() {
		return
	}

	table := tablewriter.NewTable(os.Stdout, tablewriter.WithHeaderAutoFormat(tw.Off))
	table.Header(qr.header.column_names)
	row_count := len(qr.results)
	col_count := len(qr.header.column_names)
	if len(qr.results) > 0 {
		// Convert to [][]string.
		results := make([][]string, row_count)
		for i, record := range qr.results {
			results[i] = make([]string, col_count)
			for j, elem := range record.Values() {
				results[i][j] = fmt.Sprint(elem)
			}
		}
		table.Bulk(results)
	} else {
		table.Append([]string{"No data returned."})
	}
	table.Render()

	for k, v := range qr.statistics {
		fmt.Fprintf(os.Stdout, "\n%s %f", k, v)
	}

	fmt.Fprintf(os.Stdout, "\n")
}

func (qr *QueryResult) LabelsAdded() int {
	return int(qr.getStat(LABELS_ADDED))
}

func (qr *QueryResult) NodesCreated() int {
	return int(qr.getStat(NODES_CREATED))
}

func (qr *QueryResult) NodesDeleted() int {
	return int(qr.getStat(NODES_DELETED))
}

func (qr *QueryResult) PropertiesSet() int {
	return int(qr.getStat(PROPERTIES_SET))
}

func (qr *QueryResult) RelationshipsCreated() int {
	return int(qr.getStat(RELATIONSHIPS_CREATED))
}

func (qr *QueryResult) RelationshipsDeleted() int {
	return int(qr.getStat(RELATIONSHIPS_DELETED))
}

func (qr *QueryResult) IndicesCreated() int {
	return int(qr.getStat(INDICES_CREATED))
}

func (qr *QueryResult) IndicesDeleted() int {
	return int(qr.getStat(INDICES_DELETED))
}

// Returns the query internal execution time in milliseconds
func (qr *QueryResult) InternalExecutionTime() float64 {
	return qr.getStat(INTERNAL_EXECUTION_TIME)
}

func (qr *QueryResult) CachedExecution() int {
	return int(qr.getStat(CACHED_EXECUTION))
}
