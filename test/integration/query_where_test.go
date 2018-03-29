package integration

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type test struct {
	Id int64
	A  bool
	B  bool
}

// TestWhereConditions tests adding WHERE clauses
// to a Query, ensuring it doesn't interfere with any filters.
func TestWhereConditions(t *testing.T) {
	resource, err := app.RegisterResource(test{})
	require.Nil(t, err)

	// insert some test data
	res, err := resource.InsertInstance(app.DB(), &test{A: false, B: true}).Result()
	require.Nil(t, err)
	a := res.(*test)

	res, err = resource.InsertInstance(app.DB(), &test{A: true, B: false}).Result()
	require.Nil(t, err)

	res, err = resource.InsertInstance(app.DB(), &test{A: true, B: true}).Result()
	require.Nil(t, err)

	// the following query must only return the resource with the id we're filtering by.
	//
	// before we were properly separating filters and user-defined WHERE conditions,
	// the generated WHERE clause looked like this:
	// WHERE ("test"."a") OR ("test"."b") AND (("test"."id" = 1))
	//
	// after the fix it properly generates the following WHERE clause,
	// separating filters and user-defined conditions:
	// WHERE (("test"."id" = 1)) AND (("test"."a") OR ("test"."b"))
	q := resource.Select(app.DB())
	q.Filters(resource.IdFilter(a.Id))
	q.WhereOr(`"test"."a"`)
	q.WhereOr(`"test"."b"`)
	res, err = q.Result()
	require.Nil(t, err)
	require.Len(t, res, 1)
}
