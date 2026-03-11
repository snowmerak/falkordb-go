package graph

import "context"

//go:fix inline
func (g *Graph) ExecutionPlan(query string) (string, error) {
	return g.ExecutionPlanContext(context.Background(), query)
}

//go:fix inline
func (g *Graph) Profile(query string, params map[string]interface{}, options *QueryOptions) ([]string, error) {
	return g.ProfileContext(context.Background(), query, params, options)
}

//go:fix inline
func (g *Graph) Delete() error {
	return g.DeleteContext(context.Background())
}

//go:fix inline
func (g *Graph) Query(query string, params map[string]interface{}, options *QueryOptions) (*QueryResult, error) {
	return g.QueryContext(context.Background(), query, params, options)
}

//go:fix inline
func (g *Graph) ROQuery(query string, params map[string]interface{}, options *QueryOptions) (*QueryResult, error) {
	return g.ROQueryContext(context.Background(), query, params, options)
}

//go:fix inline
func (g *Graph) Pipeline(reqs []QueryRequest) ([]*QueryResult, error) {
	return g.PipelineContext(context.Background(), reqs)
}

//go:fix inline
func (g *Graph) MemoryUsage(samples int) (map[string]interface{}, error) {
	return g.MemoryUsageContext(context.Background(), samples)
}

//go:fix inline
func (g *Graph) CallProcedure(procedure string, yield []string, args ...interface{}) (*QueryResult, error) {
	return g.CallProcedureContext(context.Background(), procedure, yield, args...)
}

//go:fix inline
func (g *Graph) SlowLog() ([]SlowLogEntry, error) {
	return g.SlowLogContext(context.Background())
}

//go:fix inline
func (g *Graph) SlowLogReset() error {
	return g.SlowLogResetContext(context.Background())
}
