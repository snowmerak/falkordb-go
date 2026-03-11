package falkordb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/snowmerak/falkordb-go/graph"
)

type FalkorDB struct {
	Conn     redis.UniversalClient
	readonly bool
}

type ConnectionOption = redis.Options

type ConnectionClusterOption = redis.ClusterOptions

func isSentinel(conn redis.UniversalClient) bool {
	if c, ok := conn.(*redis.Client); ok {
		info, _ := c.InfoMap(context.Background(), "server").Result()
		return info["Server"]["redis_mode"] == "sentinel"
	}
	return false
}

// new creates a new FalkorDB instance.
func new(options *ConnectionOption, isReadonly bool) (*FalkorDB, error) {
	db := redis.NewClient(options)

	if isSentinel(db) {
		mastersRaw, err := db.Do(context.Background(), "SENTINEL", "MASTERS").Result()
		if err != nil {
			return nil, err
		}
		masters, ok := mastersRaw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected sentinel masters type %T", mastersRaw)
		}
		if len(masters) != 1 {
			return nil, errors.New("multiple masters, require service name")
		}
		m0, ok := masters[0].(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("sentinel master entry malformed")
		}
		nameRaw, ok := m0["name"]
		if !ok {
			return nil, errors.New("sentinel master missing name")
		}
		masterName, ok := nameRaw.(string)
		if !ok {
			return nil, errors.New("sentinel master name not string")
		}
		db = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       masterName,
			SentinelAddrs:    []string{options.Addr},
			ClientName:       options.ClientName,
			Username:         options.Username,
			Password:         options.Password,
			SentinelUsername: options.Username,
			SentinelPassword: options.Password,
			MaxRetries:       options.MaxRetries,
			MinRetryBackoff:  options.MinRetryBackoff,
			MaxRetryBackoff:  options.MaxRetryBackoff,
			TLSConfig:        options.TLSConfig,
			PoolFIFO:         options.PoolFIFO,
			PoolSize:         options.PoolSize,
			PoolTimeout:      options.PoolTimeout,
		})
	}
	return &FalkorDB{
		Conn:     db,
		readonly: isReadonly,
	}, nil
}

// New creates a new FalkorDB instance.
func New(options *ConnectionOption) (*FalkorDB, error) {
	return new(options, false)
}

// NewReadOnly creates a new read-only FalkorDB instance.
func NewReadOnly(options *ConnectionOption) (*FalkorDB, error) {
	return new(options, true)
}

// NewCluster creates a new FalkorDB cluster instance.
func NewCluster(options *ConnectionClusterOption) (*FalkorDB, error) {
	db := redis.NewClusterClient(options)
	return &FalkorDB{
		Conn:     db,
		readonly: false,
	}, nil
}

// NewClusterReadOnly creates a new read-only FalkorDB cluster instance.
func NewClusterReadOnly(options *ConnectionClusterOption) (*FalkorDB, error) {
	db := redis.NewClusterClient(options)
	return &FalkorDB{
		Conn:     db,
		readonly: true,
	}, nil
}

// Creates a new FalkorDB instance from a URL.
func FromURL(url string) (*FalkorDB, error) {
	url = strings.ReplaceAll(url, "falkors://", "rediss://")
	url = strings.ReplaceAll(url, "falkor://", "redis://")

	options, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	return New(options)
}

func FromClusterURL(urls string) (*FalkorDB, error) {
	options, err := redis.ParseClusterURL(urls)
	if err != nil {
		return nil, err
	}

	return NewCluster(options)
}

// Selects a graph by creating a new Graph instance.
func (db *FalkorDB) SelectGraph(graphName string) *graph.Graph {
	return graph.NewWithMode(graphName, db.Conn, db.readonly)
}

// CopyGraph copies a graph to a new key.
func (db *FalkorDB) CopyGraph(ctx context.Context, src, dest string) error {
	return db.Conn.Do(ctx, "GRAPH.COPY", src, dest).Err()
}

// List all graph names.
// See: https://docs.falkordb.com/commands/graph.list.html
func (db *FalkorDB) ListGraphs(ctx context.Context) ([]string, error) {
	return db.Conn.Do(ctx, "GRAPH.LIST").StringSlice()
}

// Retrieve a DB level configuration.
// For a list of available configurations see: https://docs.falkordb.com/configuration.html#falkordb-configuration-parameters
func (db *FalkorDB) ConfigGet(ctx context.Context, key string) (interface{}, error) {
	return db.Conn.Do(ctx, "GRAPH.CONFIG", "GET", key).Result()
}

// Update a DB level configuration.
// For a list of available configurations see: https://docs.falkordb.com/configuration.html#falkordb-configuration-parameters
func (db *FalkorDB) ConfigSet(ctx context.Context, key string, value interface{}) error {
	return db.Conn.Do(ctx, "GRAPH.CONFIG", "SET", key, value).Err()
}

// runOnAllMasters executes a command on all master nodes if connected to a cluster.
func (db *FalkorDB) runOnAllMasters(ctx context.Context, args ...interface{}) error {
	if cc, ok := db.Conn.(*redis.ClusterClient); ok {
		return cc.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
			return client.Do(ctx, args...).Err()
		})
	}
	return db.Conn.Do(ctx, args...).Err()
}

// LoadUDF loads a user defined function library.
func (db *FalkorDB) LoadUDF(ctx context.Context, libraryName, code string) error {
	return db.runOnAllMasters(ctx, "GRAPH.UDF", "LOAD", libraryName, code)
}

// LoadUDFReplace loads a user defined function library, replacing it if it already exists.
func (db *FalkorDB) LoadUDFReplace(ctx context.Context, libraryName, code string) error {
	return db.runOnAllMasters(ctx, "GRAPH.UDF", "LOAD", "REPLACE", libraryName, code)
}

// LoadUDFFromFile loads a user defined function library from a file.
func (db *FalkorDB) LoadUDFFromFile(ctx context.Context, libraryName, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return db.LoadUDF(ctx, libraryName, string(content))
}

// LoadUDFFromFileReplace loads a user defined function library from a file, replacing it if it already exists.
func (db *FalkorDB) LoadUDFFromFileReplace(ctx context.Context, libraryName, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return db.LoadUDFReplace(ctx, libraryName, string(content))
}

// IsUdfAlreadyRegisteredError checks if the error is due to the UDF library already being registered.
func IsUdfAlreadyRegisteredError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already registered")
}

// UDFLibrary represents a user defined function library.
type UDFLibrary struct {
	Name      string
	Functions []string
	Code      string
}

type UDFListOptions struct {
	LibraryName string
	WithCode    bool
}

type UDFListOption func(*UDFListOptions)

// WithUDFLibrary filters the list by library name.
func WithUDFLibrary(name string) UDFListOption {
	return func(o *UDFListOptions) {
		o.LibraryName = name
	}
}

// WithUDFCode includes the source code in the response.
func WithUDFCode() UDFListOption {
	return func(o *UDFListOptions) {
		o.WithCode = true
	}
}

// ListUDF lists loaded user defined function libraries.
// It accepts optional arguments to filter by library name and to include source code.
func (db *FalkorDB) ListUDF(ctx context.Context, opts ...UDFListOption) ([]UDFLibrary, error) {
	options := &UDFListOptions{}
	for _, opt := range opts {
		opt(options)
	}

	args := []interface{}{"LIST"}
	if options.LibraryName != "" {
		args = append(args, options.LibraryName)
	}
	if options.WithCode {
		args = append(args, "WITHCODE")
	}

	cmdArgs := append([]interface{}{"GRAPH.UDF"}, args...)
	res, err := db.Conn.Do(ctx, cmdArgs...).Result()
	if err != nil {
		return nil, err
	}

	rawList, ok := res.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type from GRAPH.UDF LIST: %T", res)
	}

	libraries := make([]UDFLibrary, 0, len(rawList))
	for _, item := range rawList {
		rawLib, ok := item.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected item type in GRAPH.UDF LIST response: %T", item)
		}

		lib := UDFLibrary{}

		if name, ok := rawLib["library_name"].(string); ok {
			lib.Name = name
		}

		if funcs, ok := rawLib["functions"].([]interface{}); ok {
			lib.Functions = make([]string, len(funcs))
			for i, f := range funcs {
				if s, ok := f.(string); ok {
					lib.Functions[i] = s
				}
			}
		}

		if code, ok := rawLib["library_code"].(string); ok {
			lib.Code = code
		}

		libraries = append(libraries, lib)
	}

	return libraries, nil
}

// DeleteUDF removes a user defined function library.
func (db *FalkorDB) DeleteUDF(ctx context.Context, libraryName string) error {
	return db.runOnAllMasters(ctx, "GRAPH.UDF", "DELETE", libraryName)
}

// FlushUDFs removes all user defined function libraries.
func (db *FalkorDB) FlushUDFs(ctx context.Context) error {
	return db.runOnAllMasters(ctx, "GRAPH.UDF", "FLUSH")
}

// CreateConstraint creates a constraint on the graph.
// constraintType: "MANDATORY" or "UNIQUE".
// entityType: "NODE" or "RELATIONSHIP".
// label: the node label or relationship type.
// properties: attribute names to constrain.
// See: https://docs.falkordb.com/commands/graph.constraint-create.html
func (db *FalkorDB) CreateConstraint(ctx context.Context, graphName, constraintType, entityType, label string, properties []string) error {
	args := []interface{}{
		"GRAPH.CONSTRAINT", "CREATE",
		graphName, constraintType, entityType, label,
		"PROPERTIES", len(properties),
	}
	for _, p := range properties {
		args = append(args, p)
	}
	return db.Conn.Do(ctx, args...).Err()
}

// DropConstraint removes a constraint from the graph.
// See: https://docs.falkordb.com/commands/graph.constraint-drop.html
func (db *FalkorDB) DropConstraint(ctx context.Context, graphName, constraintType, entityType, label string, properties []string) error {
	args := []interface{}{
		"GRAPH.CONSTRAINT", "DROP",
		graphName, constraintType, entityType, label,
		"PROPERTIES", len(properties),
	}
	for _, p := range properties {
		args = append(args, p)
	}
	return db.Conn.Do(ctx, args...).Err()
}

// InfoSection specifies which queries to retrieve from GRAPH.INFO.
type InfoSection string

const (
	// InfoAll returns both running and waiting queries.
	InfoAll InfoSection = ""
	// InfoRunningQueries returns only running queries.
	InfoRunningQueries InfoSection = "RunningQueries"
	// InfoWaitingQueries returns only waiting queries.
	InfoWaitingQueries InfoSection = "WaitingQueries"
)

// GraphInfoQuery represents a single query entry returned by GRAPH.INFO.
type GraphInfoQuery struct {
	ReceivedAt int64
	Graph      string
	Query      string
}

// GraphInfo contains the running and waiting queries returned by GRAPH.INFO.
type GraphInfo struct {
	RunningQueries []GraphInfoQuery
	WaitingQueries []GraphInfoQuery
}

// Info returns information about running and/or waiting queries.
// See: https://docs.falkordb.com/commands/graph.info.html
func (db *FalkorDB) Info(ctx context.Context, section InfoSection) (*GraphInfo, error) {
	args := []interface{}{"GRAPH.INFO"}
	if section != InfoAll {
		args = append(args, string(section))
	}

	res, err := db.Conn.Do(ctx, args...).Result()
	if err != nil {
		return nil, err
	}

	raw, ok := res.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected GRAPH.INFO response type %T", res)
	}

	info := &GraphInfo{}

	for i := 0; i < len(raw)-1; i += 2 {
		header, ok := raw[i].(string)
		if !ok {
			continue
		}
		queries, ok := raw[i+1].([]interface{})
		if !ok {
			continue
		}

		parsed := make([]GraphInfoQuery, 0, len(queries))
		for _, q := range queries {
			entry, ok := q.([]interface{})
			if !ok {
				continue
			}
			giq := GraphInfoQuery{}
			if len(entry) > 0 {
				if v, ok := entry[0].(int64); ok {
					giq.ReceivedAt = v
				}
			}
			if len(entry) > 1 {
				if v, ok := entry[1].(string); ok {
					giq.Graph = v
				}
			}
			if len(entry) > 2 {
				if v, ok := entry[2].(string); ok {
					giq.Query = v
				}
			}
			parsed = append(parsed, giq)
		}

		if strings.Contains(header, "Running") {
			info.RunningQueries = parsed
		} else if strings.Contains(header, "Waiting") {
			info.WaitingQueries = parsed
		}
	}

	return info, nil
}
