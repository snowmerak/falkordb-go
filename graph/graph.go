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
