[![license](https://img.shields.io/github/license/snowmerak/falkordb-go.svg)](https://github.com/snowmerak/falkordb-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/snowmerak/falkordb-go)](https://goreportcard.com/report/github.com/snowmerak/falkordb-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/snowmerak/falkordb-go.svg)](https://pkg.go.dev/github.com/snowmerak/falkordb-go)

# falkordb-go

`falkordb-go` is a Golang client for the [FalkorDB](https://falkordb.com) database.

## Overview
- Lightweight client with simple Query/ROQuery APIs.
- Parses nodes, edges, paths, arrays, maps, points, and vectors into Go types.
- Exposes query statistics plus PrettyPrint for quick inspection.
- Supports single instance, cluster, sentinel discovery, and TLS via URL schemes.
- `trunk` is the primary, up-to-date branch.

## Quick start
1) Start a local FalkorDB (standalone):

```
docker compose -f docker-compose.standalone.yml up -d
```

2) Add the client to your module:

```
go get github.com/snowmerak/falkordb-go
```

3) Run a minimal query:

```go
package main

import (
    "log"

    "github.com/snowmerak/falkordb-go"
    "github.com/snowmerak/falkordb-go/graph"
)

func main() {
    db, err := falkordb.FromURL("falkor://0.0.0.0:6379")
    if err != nil {
        log.Fatal(err)
    }

    g := db.SelectGraph("social")

    _, err = g.Query("CREATE (:Person {name:'John Doe', age:33})", nil, nil)
    if err != nil {
        log.Fatal(err)
    }

    opts := graph.NewQueryOptions().SetTimeout(10) // ms timeout
    res, err := g.Query("MATCH (p:Person) RETURN p.name, p.age", nil, opts)
    if err != nil {
        log.Fatal(err)
    }

    res.PrettyPrint()
}
```

## Usage and examples

The complete API is documented on [pkg.go.dev](https://pkg.go.dev/github.com/snowmerak/falkordb-go).

- Query vs ROQuery

```go
res, err := g.Query("MATCH (p:Person) RETURN p.name", nil, nil)
roRes, err := g.ROQuery("MATCH (p:Person) RETURN p.name", nil, nil)
```

- Iterating results

```go
for res.Next() {
    r := res.Record()
    name := r.GetByIndex(0)
    log.Printf("name=%v", name)
}
```

- With timeouts (milliseconds)

```go
opts := graph.NewQueryOptions().SetTimeout(5)
res, err := g.Query("UNWIND range(0, 1000000) AS v RETURN v", nil, opts)
```

- Read-only client

```go
db, err := falkordb.NewReadOnly(&falkordb.ConnectionOption{Addr: "0.0.0.0:6379"})
if err != nil { log.Fatal(err) }

g := db.SelectGraph("social")

// This will error because the graph is read-only
_, err = g.Query("CREATE (:X)", nil, nil)

// RO queries are allowed
res, err := g.ROQuery("MATCH (n) RETURN n", nil, nil)
```

- Pipelined batch queries

```go
reqs := []graph.QueryRequest{
    { // defaults to GRAPH.QUERY when Command is empty
        Query:   "MATCH (p:Person) RETURN p",
        Options: graph.NewQueryOptions().SetTimeout(50),
    },
    {
        Command: graph.CmdROQuery, // or graph.CmdQuery / empty for write queries
        Query:   "MATCH (c:Country {name:$name}) RETURN c",
        Params:  map[string]interface{}{"name": "Japan"},
    },
}
batch, err := g.Pipeline(reqs)
if err != nil {
    log.Fatal(err)
}
// batch[0], batch[1] are ordered results
```

## Running queries with timeouts

Queries can be run with a millisecond-level timeout as described in [the documentation](https://docs.falkordb.com/configuration.html#timeout). To take advantage of this feature, the `QueryOptions` struct should be used:

```go
options := graph.NewQueryOptions().SetTimeout(10) // 10-millisecond timeout
res, err := g.Query("MATCH (src {name: 'John Doe'})-[*]->(dest) RETURN dest", nil, options)
```

## Advanced Graph Operations

### Profile

You can profile a query execution plan using the `Profile` method.

```go
res, err := g.Profile("MATCH (p:Person) RETURN p", nil, nil)
if err != nil {
    log.Fatal(err)
}
res.PrettyPrint() // Prints the execution plan
```

### Copy Graph

You can copy a graph to a new key.

```go
err := db.CopyGraph("social", "social_backup")
```

### Memory Usage

You can retrieve the memory usage of a specific graph.

```go
mem, err := g.MemoryUsage()
// mem is a map[string]interface{} containing memory stats
```

## User Defined Functions (UDFs)

`falkordb-go` supports managing UDF libraries.

### Loading UDFs

You can load UDFs from a string or a file. You can also use the `Replace` variants to overwrite existing libraries.

```go
// Load from string
err := db.LoadUDF("mylib", "def my_func(a, b): return a + b")

// Load from file
err := db.LoadUDFFromFile("mylib", "/path/to/lib.py")

// Load and replace if exists
err := db.LoadUDFReplace("mylib", "def my_func(a, b): return a * b")
err := db.LoadUDFFromFileReplace("mylib", "/path/to/lib.py")
```

### Listing UDFs

You can list loaded UDF libraries, optionally filtering by name or including the source code.

```go
// List all libraries
libs, err := db.ListUDF()

// List specific library
libs, err := db.ListUDF(falkordb.WithUDFLibrary("mylib"))

// List with source code
libs, err := db.ListUDF(falkordb.WithUDFCode())
```

### Deleting UDFs

You can delete a specific library or flush all libraries.

```go
// Delete a specific library
err := db.DeleteUDF("mylib")

// Flush all libraries
err := db.FlushUDFs()
```

## Supported Types

`falkordb-go` automatically maps FalkorDB types to Go types:

- **Nodes**: `graph.Node`
- **Edges**: `graph.Edge`
- **Paths**: `graph.Path`
- **Maps**: `map[string]interface{}`
- **Arrays**: `[]interface{}`
- **Integers/Floats**: `int64`, `float64`
- **Strings**: `string`
- **Booleans**: `bool`
- **Null**: `nil`
- **Spatial Types**: `graph.Point`
- **Vector Types**: `[]float32`
- **Date/Time Types**:
    - `date`: `time.Time`
    - `localtime`: `time.Time` (Year 0)
    - `localdatetime`: `time.Time`
    - `duration`: `time.Duration`

## Connection options
- Single instance: `falkordb.FalkorDBNew(&falkordb.ConnectionOption{Addr: "0.0.0.0:6379"})`
- Cluster: `falkordb.FalkorDBNewCluster(&falkordb.ConnectionClusterOption{Addrs: []string{"0.0.0.0:6379"}})`
- URL-based (sentinel/TLS aware): `falkordb.FromURL("falkor://host:port")` or `falkors://` for TLS.
- Environment defaults used in tests: `FALKORDB_ADDR` for host:port, `FALKORDB_TEST_MODE=cluster` to switch client mode.

## Examples
- Start a standalone server: `docker compose -f docker-compose.standalone.yml up -d` or `task standalone:up`
- Start a clustered server: `docker compose -f docker-compose.cluster.yml up -d` or `task cluster:down`

## Running tests

A simple test suite is provided, and can be run with:

```
task test
# or
go test ./...
```

The tests expect a FalkorDB server to be available at localhost:6379 (or the address in `FALKORDB_ADDR`). Task automation is defined in [`Taskfile.yml`](https://taskfile.dev).

## License

falkordb-go is distributed under the BSD3 license - see [LICENSE](LICENSE)