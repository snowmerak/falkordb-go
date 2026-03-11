package integration_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstraintUnique(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	graphName := "constraint_test_unique"
	g := db.SelectGraph(graphName)
	defer g.Delete()

	// Create some data first
	_, err := g.Query("CREATE (:Person {first_name: 'John', last_name: 'Doe'})", nil, nil)
	assert.Nil(t, err)

	// Create index (required before UNIQUE constraint)
	_, err = g.Query("CREATE INDEX FOR (p:Person) ON (p.first_name, p.last_name)", nil, nil)
	assert.Nil(t, err)

	// Wait for index to be created
	time.Sleep(500 * time.Millisecond)

	// Create UNIQUE constraint
	err = db.CreateConstraint(graphName, "UNIQUE", "NODE", "Person", []string{"first_name", "last_name"})
	assert.Nil(t, err)

	// Wait for async constraint creation
	time.Sleep(1 * time.Second)

	// Try to create a duplicate — should fail
	_, err = g.Query("CREATE (:Person {first_name: 'John', last_name: 'Doe'})", nil, nil)
	assert.NotNil(t, err, "Creating a duplicate should fail due to unique constraint")

	// Drop constraint
	err = db.DropConstraint(graphName, "UNIQUE", "NODE", "Person", []string{"first_name", "last_name"})
	assert.Nil(t, err)

	// Now creating a duplicate should succeed
	_, err = g.Query("CREATE (:Person {first_name: 'John', last_name: 'Doe'})", nil, nil)
	assert.Nil(t, err, "Creating a duplicate should succeed after constraint is dropped")
}

func TestConstraintMandatory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	graphName := "constraint_test_mandatory"
	g := db.SelectGraph(graphName)
	defer g.Delete()

	// Create some data with the required attribute
	_, err := g.Query("CREATE (:Employee {emp_id: 1, name: 'Alice'})", nil, nil)
	assert.Nil(t, err)

	// Create MANDATORY constraint on emp_id
	err = db.CreateConstraint(graphName, "MANDATORY", "NODE", "Employee", []string{"emp_id"})
	assert.Nil(t, err)

	// Wait for async constraint creation
	time.Sleep(1 * time.Second)

	// Try to create a node without the mandatory property — should fail
	_, err = g.Query("CREATE (:Employee {name: 'Bob'})", nil, nil)
	assert.NotNil(t, err, "Creating a node without mandatory 'emp_id' should fail")

	// Drop constraint
	err = db.DropConstraint(graphName, "MANDATORY", "NODE", "Employee", []string{"emp_id"})
	assert.Nil(t, err)

	// Now creating a node without emp_id should succeed
	_, err = g.Query("CREATE (:Employee {name: 'Charlie'})", nil, nil)
	assert.Nil(t, err, "Creating a node without emp_id should succeed after constraint is dropped")
}
