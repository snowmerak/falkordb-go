package graph

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/snowmerak/falkordb-go/util/strs"
)

const (
	CmdQuery   = "GRAPH.QUERY"
	CmdROQuery = "GRAPH.RO_QUERY"
	CmdProfile = "GRAPH.PROFILE"
)

var ctx = context.Background()

// QueryOptions are a set of additional arguments to be emitted with a query.
type QueryOptions struct {
	timeout int
}

// QueryRequest represents a single graph command to enqueue in a pipeline.
type QueryRequest struct {
	Command string
	Query   string
	Params  map[string]interface{}
	Options *QueryOptions
}

// Graph represents a graph, which is a collection of nodes and edges.
type Graph struct {
	Id       string
	Conn     redis.UniversalClient
	schema   GraphSchema
	readonly bool
}

// New creates a new graph.
func New(Id string, conn redis.UniversalClient) *Graph {
	return NewWithMode(Id, conn, false)
}

// NewWithMode creates a new graph with explicit read-only mode.
func NewWithMode(Id string, conn redis.UniversalClient, readonly bool) *Graph {
	g := new(Graph)
	g.Id = Id
	g.Conn = conn
	g.readonly = readonly
	g.schema = GraphSchemaNew(g)
	return g
}

// NewGraphWithSchema creates a graph instance seeded with an existing schema (used in tests).
func NewGraphWithSchema(schema GraphSchema) *Graph {
	return &Graph{schema: schema}
}

// ExecutionPlan gets the execution plan for given query.
func (g *Graph) ExecutionPlan(query string) (string, error) {
	return g.Conn.Do(ctx, "GRAPH.EXPLAIN", g.Id, query).Text()
}

// Profile executes a query and returns an execution plan augmented with metrics.
func (g *Graph) Profile(query string, params map[string]interface{}, options *QueryOptions) ([]string, error) {
	if params != nil {
		query = BuildParamsHeader(params) + query
	}

	args := []interface{}{g.Id, query, "--compact"}
	if options != nil && options.timeout >= 0 {
		args = append(args, "timeout", options.timeout)
	}

	cmdArgs := append([]interface{}{CmdProfile}, args...)
	res, err := g.Conn.Do(ctx, cmdArgs...).Result()
	if err != nil {
		return nil, err
	}

	return parseProfileResponse(res)
}

func parseProfileResponse(res interface{}) ([]string, error) {
	raw, ok := res.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected profile response type %T", res)
	}

	lines := make([]string, len(raw))
	for i, r := range raw {
		s, ok := r.(string)
		if !ok {
			return nil, fmt.Errorf("profile entry %d not string", i)
		}
		lines[i] = strings.TrimSpace(s)
	}
	return lines, nil
}

// Delete removes the graph.
func (g *Graph) Delete() error {
	err := g.Conn.Do(ctx, "GRAPH.DELETE", g.Id).Err()

	// clear internal mappings
	g.schema.clear()

	return err
}

// NewQueryOptions instantiates a new QueryOptions struct.
func NewQueryOptions() *QueryOptions {
	return &QueryOptions{
		timeout: -1,
	}
}

// SetTimeout sets the timeout member of the QueryOptions struct
func (options *QueryOptions) SetTimeout(timeout int) *QueryOptions {
	options.timeout = timeout
	return options
}

// GetTimeout retrieves the timeout of the QueryOptions struct
func (options *QueryOptions) GetTimeout() int {
	return options.timeout
}

func (g *Graph) query(command string, query string, params map[string]interface{}, options *QueryOptions) (*QueryResult, error) {
	if g.readonly && command != CmdROQuery {
		return nil, errors.New("graph is read-only")
	}
	if params != nil {
		query = BuildParamsHeader(params) + query
	}

	args := []interface{}{g.Id, query, "--compact"}
	if options != nil && options.timeout >= 0 {
		args = append(args, "timeout", options.timeout)
	}

	cmdArgs := append([]interface{}{command}, args...)
	r, err := g.Conn.Do(ctx, cmdArgs...).Result()
	if err != nil {
		return nil, err
	}

	return QueryResultNew(g, r)
}

// Pipeline executes multiple graph commands in a single round-trip and returns results in order.
// Each request can target GRAPH.QUERY or GRAPH.RO_QUERY via the Command field (defaults to GRAPH.QUERY).
func (g *Graph) Pipeline(reqs []QueryRequest) ([]*QueryResult, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	pipe := g.Conn.Pipeline()
	cmds := make([]*redis.Cmd, len(reqs))

	for i, req := range reqs {
		q := req.Query
		if req.Params != nil {
			q = BuildParamsHeader(req.Params) + q
		}

		args := []interface{}{g.Id, q, "--compact"}
		if req.Options != nil && req.Options.timeout >= 0 {
			args = append(args, "timeout", req.Options.timeout)
		}

		command := req.Command
		if command == "" {
			command = CmdQuery
		}
		if g.readonly && command != CmdROQuery {
			return nil, errors.New("graph is read-only")
		}

		cmdArgs := append([]interface{}{command}, args...)
		cmds[i] = pipe.Do(ctx, cmdArgs...)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	results := make([]*QueryResult, len(reqs))
	for i, cmd := range cmds {
		if err := cmd.Err(); err != nil {
			return nil, err
		}
		r := cmd.Val()
		qr, err := QueryResultNew(g, r)
		if err != nil {
			return nil, err
		}
		results[i] = qr
	}

	return results, nil
}

// Query executes a query against the graph.
func (g *Graph) Query(query string, params map[string]interface{}, options *QueryOptions) (*QueryResult, error) {
	return g.query(CmdQuery, query, params, options)
}

// ROQuery executes a read only query against the graph.
func (g *Graph) ROQuery(query string, params map[string]interface{}, options *QueryOptions) (*QueryResult, error) {
	return g.query(CmdROQuery, query, params, options)
}

// Procedures

// MemoryUsage returns detailed memory consumption statistics for a specific graph.
// samples: Number of samples to take when estimating memory usage. (default 100 if -1)
func (g *Graph) MemoryUsage(samples int) (map[string]interface{}, error) {
	args := []interface{}{"GRAPH.MEMORY", "USAGE", g.Id}
	if samples > 0 {
		args = append(args, "SAMPLES", samples)
	}

	res, err := g.Conn.Do(ctx, args...).Result()
	if err != nil {
		return nil, err
	}

	// Try parsing as array
	if rawArray, ok := res.([]interface{}); ok {
		// Response is array of key-value pairs
		// [key1, value1, key2, value2, ...]
		memoryInfo := make(map[string]interface{})
		for i := 0; i < len(rawArray); i += 2 {
			key, ok := rawArray[i].(string)
			if !ok {
				return nil, fmt.Errorf("memory info key at index %d is not string", i)
			}

			val := rawArray[i+1]
			memoryInfo[key] = val
		}
		return memoryInfo, nil
	}

	// Try parsing as map (some go-redis client versions might parse it automatically or if the command returns a map)
	if rawMap, ok := res.(map[interface{}]interface{}); ok {
		memoryInfo := make(map[string]interface{})
		for k, v := range rawMap {
			keyStr, ok := k.(string)
			if !ok {
				continue // Or handle error
			}
			memoryInfo[keyStr] = v
		}
		return memoryInfo, nil
	}

	return nil, fmt.Errorf("unexpected memory response type %T", res)
}

// CallProcedure invokes procedure.
func (g *Graph) CallProcedure(procedure string, yield []string, args ...interface{}) (*QueryResult, error) {
	query := fmt.Sprintf("CALL %s(", procedure)

	tmp := make([]string, 0, len(args))
	for arg := range args {
		tmp = append(tmp, strs.ToString(arg))
	}
	query += fmt.Sprintf("%s)", strings.Join(tmp, ","))

	if len(yield) > 0 {
		query += fmt.Sprintf(" YIELD %s", strings.Join(yield, ","))
	}

	return g.Query(query, nil, nil)
}

// BuildParamsHeader builds a CYPHER params header from key/value pairs.
func BuildParamsHeader(params map[string]interface{}) string {
	header := "CYPHER "
	for key, value := range params {
		header += fmt.Sprintf("%s=%v ", key, strs.ToString(value))
	}
	return header
}
