package falkordb

import "context"

//go:fix inline
func (db *FalkorDB) CopyGraph(src, dest string) error {
	return db.CopyGraphContext(context.Background(), src, dest)
}

//go:fix inline
func (db *FalkorDB) ListGraphs() ([]string, error) {
	return db.ListGraphsContext(context.Background())
}

//go:fix inline
func (db *FalkorDB) ConfigGet(key string) (interface{}, error) {
	return db.ConfigGetContext(context.Background(), key)
}

//go:fix inline
func (db *FalkorDB) ConfigSet(key string, value interface{}) error {
	return db.ConfigSetContext(context.Background(), key, value)
}

func (db *FalkorDB) runOnAllMasters(args ...interface{}) error {
	return db.runOnAllMastersContext(context.Background(), args...)
}

//go:fix inline
func (db *FalkorDB) LoadUDF(libraryName, code string) error {
	return db.LoadUDFContext(context.Background(), libraryName, code)
}

//go:fix inline
func (db *FalkorDB) LoadUDFReplace(libraryName, code string) error {
	return db.LoadUDFReplaceContext(context.Background(), libraryName, code)
}

//go:fix inline
func (db *FalkorDB) LoadUDFFromFile(libraryName, filePath string) error {
	return db.LoadUDFFromFileContext(context.Background(), libraryName, filePath)
}

//go:fix inline
func (db *FalkorDB) LoadUDFFromFileReplace(libraryName, filePath string) error {
	return db.LoadUDFFromFileReplaceContext(context.Background(), libraryName, filePath)
}

//go:fix inline
func (db *FalkorDB) ListUDF(opts ...UDFListOption) ([]UDFLibrary, error) {
	return db.ListUDFContext(context.Background(), opts...)
}

//go:fix inline
func (db *FalkorDB) DeleteUDF(libraryName string) error {
	return db.DeleteUDFContext(context.Background(), libraryName)
}

//go:fix inline
func (db *FalkorDB) FlushUDFs() error {
	return db.FlushUDFsContext(context.Background())
}

//go:fix inline
func (db *FalkorDB) CreateConstraint(graphName, constraintType, entityType, label string, properties []string) error {
	return db.CreateConstraintContext(context.Background(), graphName, constraintType, entityType, label, properties)
}

//go:fix inline
func (db *FalkorDB) DropConstraint(graphName, constraintType, entityType, label string, properties []string) error {
	return db.DropConstraintContext(context.Background(), graphName, constraintType, entityType, label, properties)
}

//go:fix inline
func (db *FalkorDB) Info(section InfoSection) (*GraphInfo, error) {
	return db.InfoContext(context.Background(), section)
}
