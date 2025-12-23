package falkordb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/snowmerak/falkordb-go/graph"
)

var ctx = context.Background()

type FalkorDB struct {
	Conn redis.UniversalClient
}

type ConnectionOption = redis.Options

type ConnectionClusterOption = redis.ClusterOptions

func isSentinel(conn redis.UniversalClient) bool {
	if c, ok := conn.(*redis.Client); ok {
		info, _ := c.InfoMap(ctx, "server").Result()
		return info["Server"]["redis_mode"] == "sentinel"
	}
	return false
}

// New creates a new FalkorDB instance.
func New(options *ConnectionOption) (*FalkorDB, error) {
	db := redis.NewClient(options)

	if isSentinel(db) {
		mastersRaw, err := db.Do(ctx, "SENTINEL", "MASTERS").Result()
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
	return &FalkorDB{Conn: db}, nil
}

// NewCluster creates a new FalkorDB cluster instance.
func NewCluster(options *ConnectionClusterOption) (*FalkorDB, error) {
	db := redis.NewClusterClient(options)
	return &FalkorDB{Conn: db}, nil
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
	return graph.New(graphName, db.Conn)
}

// List all graph names.
// See: https://docs.falkordb.com/commands/graph.list.html
func (db *FalkorDB) ListGraphs() ([]string, error) {
	return db.Conn.Do(ctx, "GRAPH.LIST").StringSlice()
}

// Retrieve a DB level configuration.
// For a list of available configurations see: https://docs.falkordb.com/configuration.html#falkordb-configuration-parameters
func (db *FalkorDB) ConfigGet(key string) (interface{}, error) {
	return db.Conn.Do(ctx, "GRAPH.CONFIG", "GET", key).Result()
}

// Update a DB level configuration.
// For a list of available configurations see: https://docs.falkordb.com/configuration.html#falkordb-configuration-parameters
func (db *FalkorDB) ConfigSet(key string, value interface{}) error {
	return db.Conn.Do(ctx, "GRAPH.CONFIG", "SET", key, value).Err()
}
