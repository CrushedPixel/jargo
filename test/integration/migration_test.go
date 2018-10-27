// +build integration

package integration

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type MigrationTestType0 struct {
	Id   int64 `jargo:",table:migration_test_types"`
	Name string
	Age  int
}

type MigrationTestType1 struct {
	Id    int64 `jargo:",table:migration_test_types"`
	Name  string
	Age   int
	Valid bool `jargo:",default:TRUE"`
}

func TestMigrations(t *testing.T) {
	// register resource with name "migration_test_type"
	resource0, err := app.RegisterResource(MigrationTestType0{})
	require.Nil(t, err)
	_, err = resource0.InsertInstance(app.DB(), &MigrationTestType0{
		Name: "Peter",
		Age:  32,
	}).Result()
	require.Nil(t, err)

	// register another resource with the same name,
	// but an additional field
	resource1, err := app.RegisterResource(MigrationTestType1{})
	require.Nil(t, err)

	// the valid field should have been added by the migration,
	// and set to true by the default option.
	res, err := resource1.Select(app.DB()).
		Where(`valid = TRUE`).
		Result()
	require.Nil(t, err)
	results := res.([]*MigrationTestType1)
	require.Len(t, results, 1)
}
