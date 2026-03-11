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
- Full `context.Context` support via `*Context` methods.
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
    "context"
    "log"

    "github.com/snowmerak/falkordb-go"
    "github.com/snowmerak/falkordb-go/graph"
)

func main() {
    ctx := context.Background()

    db, err := falkordb.FromURL("falkor://0.0.0.0:6379")
    if err != nil {
        log.Fatal(err)
    }

    g := db.SelectGraph("social")

    _, err = g.QueryContext(ctx, "CREATE (:Person {name:'John Doe', age:33})", nil, nil)
    if err != nil {
        log.Fatal(err)
    }

    opts := graph.NewQueryOptions().SetTimeout(10) // ms timeout
    res, err := g.QueryContext(ctx, "MATCH (p:Person) RETURN p.name, p.age", nil, opts)
    if err != nil {
        log.Fatal(err)
    }

    res.PrettyPrint()
}
```

## Migration from v1

All methods without `Context` suffix are deprecated and annotated with `//go:fix inline`. Run `go fix ./...` to automatically migrate:

```diff
-res, err := g.Query("MATCH (n) RETURN n", nil, nil)
+res, err := g.QueryContext(context.Background(), "MATCH (n) RETURN n", nil, nil)
```

The old methods remain available for backward compatibility but will delegate to `context.Background()` internally.

## Usage and examples

The complete API is documented on [pkg.go.dev](https://pkg.go.dev/github.com/snowmerak/falkordb-go).

- Query vs ROQuery

```go
res, err := g.QueryContext(ctx, "MATCH (p:Person) RETURN p.name", nil, nil)
roRes, err := g.ROQueryContext(ctx, "MATCH (p:Person) RETURN p.name", nil, nil)
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
res, err := g.QueryContext(ctx, "UNWIND range(0, 1000000) AS v RETURN v", nil, opts)
```

- Read-only client

```go
db, err := falkordb.NewReadOnly(&falkordb.ConnectionOption{Addr: "0.0.0.0:6379"})
if err != nil { log.Fatal(err) }

g := db.SelectGraph("social")

// This will error because the graph is read-only
_, err = g.QueryContext(ctx, "CREATE (:X)", nil, nil)

// RO queries are allowed
res, err := g.ROQueryContext(ctx, "MATCH (n) RETURN n", nil, nil)
```

- Pipelined batch queries

```go
reqs := []graph.QueryRequest{
    {
        Query:   "MATCH (p:Person) RETURN p",
        Options: graph.NewQueryOptions().SetTimeout(50),
    },
    {
        Command: graph.CmdROQuery,
        Query:   "MATCH (c:Country {name:$name}) RETURN c",
        Params:  map[string]interface{}{"name": "Japan"},
    },
}
batch, err := g.PipelineContext(ctx, reqs)
if err != nil {
    log.Fatal(err)
}
// batch[0], batch[1] are ordered results
```

## Running queries with timeouts

```go
options := graph.NewQueryOptions().SetTimeout(10) // 10-millisecond timeout
res, err := g.QueryContext(ctx, "MATCH (src {name: 'John Doe'})-[*]->(dest) RETURN dest", nil, options)
```

## Advanced Graph Operations

### Profile

```go
lines, err := g.ProfileContext(ctx, "MATCH (p:Person) RETURN p", nil, nil)
// lines is []string with execution plan
```

### Copy Graph

```go
err := db.CopyGraphContext(ctx, "social", "social_backup")
```

### Memory Usage

```go
mem, err := g.MemoryUsageContext(ctx, -1) // -1 for default sample count
```

### Constraints

```go
// Create a UNIQUE constraint (requires an existing index)
err := db.CreateConstraintContext(ctx, "social", "UNIQUE", "NODE", "Person", []string{"name"})

// Create a MANDATORY constraint
err := db.CreateConstraintContext(ctx, "social", "MANDATORY", "NODE", "Person", []string{"age"})

// Drop a constraint
err := db.DropConstraintContext(ctx, "social", "MANDATORY", "NODE", "Person", []string{"age"})
```

### Slow Log

```go
entries, err := g.SlowLogContext(ctx)
for _, e := range entries {
    log.Printf("ts=%s cmd=%s query=%s duration=%s", e.Timestamp, e.Command, e.Query, e.Duration)
}

err = g.SlowLogResetContext(ctx)
```

### Server Info

```go
info, err := db.InfoContext(ctx, falkordb.InfoAll) // or InfoRunningQueries, InfoWaitingQueries
log.Printf("Running: %d, Waiting: %d", len(info.RunningQueries), len(info.WaitingQueries))
```

## User Defined Functions (UDFs)

### Loading UDFs

```go
err := db.LoadUDFContext(ctx, "mylib", "function my_func(a, b) { return a + b }")

err := db.LoadUDFFromFileContext(ctx, "mylib", "/path/to/lib.js")

// Load and replace if exists
err := db.LoadUDFReplaceContext(ctx, "mylib", "function my_func(a, b) { return a * b }")
```

### Listing UDFs

```go
libs, err := db.ListUDFContext(ctx)

libs, err := db.ListUDFContext(ctx, falkordb.WithUDFLibrary("mylib"))

libs, err := db.ListUDFContext(ctx, falkordb.WithUDFCode())
```

### Deleting UDFs

```go
err := db.DeleteUDFContext(ctx, "mylib")

err := db.FlushUDFsContext(ctx)
```

## Supported Types

| FalkorDB Type | Go Type |
|---|---|
| Node | `domain.Node` |
| Edge | `domain.Edge` |
| Path | `domain.Path` |
| Map | `map[string]interface{}` |
| Array | `[]interface{}` |
| Integer | `int64` |
| Float | `float64` |
| String | `string` |
| Boolean | `bool` |
| Null | `nil` |
| Point | `map[string]interface{}` |
| Vector | `[]float32` |
| Date / LocalDateTime | `time.Time` |
| LocalTime | `time.Time` |
| Duration | `time.Duration` |

## Connection options
- Single instance: `falkordb.New(&falkordb.ConnectionOption{Addr: "0.0.0.0:6379"})`
- Cluster: `falkordb.NewCluster(&falkordb.ConnectionClusterOption{Addrs: []string{"0.0.0.0:6379"}})`
- URL-based (sentinel/TLS aware): `falkordb.FromURL("falkor://host:port")` or `falkors://` for TLS.

## Running tests

```
task test
# or
go test ./...
```

The tests expect a FalkorDB server at localhost:6379 (or `FALKORDB_ADDR`). Task automation is in [`Taskfile.yml`](https://taskfile.dev).

## License

falkordb-go is distributed under the BSD3 license - see [LICENSE](LICENSE)