package integration

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type oneToManyA struct {
	Id   int64
	Attr string
	Bs   []oneToManyB `jargo:",has:A"`
}

type oneToManyB struct {
	Id int64
	A  oneToManyA `jargo:",belongsTo"`
}

// TestOneToManyRelations tests the behaviour of one-to-many relations.
func TestOneToManyRelations(t *testing.T) {
	resourceA, err := app.RegisterResource(oneToManyA{})
	require.Nil(t, err)

	resourceB, err := app.RegisterResource(oneToManyB{})
	require.Nil(t, err)

	// create instance of oneToManyA
	res, err := resourceA.InsertInstance(app.DB(), &oneToManyA{Attr: "test"}).Result()
	require.Nil(t, err)
	a := res.(*oneToManyA)

	// create instance of oneToManyB with relation to a
	res, err = resourceB.InsertInstance(app.DB(), &oneToManyB{A: *a}).Result()
	require.Nil(t, err)
	b := res.(*oneToManyB)

	// ensure relation is properly set
	require.Equal(t, a.Id, b.A.Id)
	require.Equal(t, a.Attr, b.A.Attr)

	// fetch oneToManyA to update relations
	res, err = resourceA.SelectById(app.DB(), a.Id).Result()
	require.Nil(t, err)
	a = res.(*oneToManyA)

	// ensure relation is properly set
	require.Equal(t, b.Id, a.Bs[0].Id)
}

type oneToOneA struct {
	Id   int64
	Attr string
	B    *oneToOneB `jargo:",has:A"`
}

type oneToOneB struct {
	Id int64
	A  oneToOneA `jargo:",belongsTo"`
}

// TestOneToOneRelations tests the behaviour of one-to-one relations.
func TestOneToOneRelations(t *testing.T) {
	resourceA, err := app.RegisterResource(oneToOneA{})
	require.Nil(t, err)

	resourceB, err := app.RegisterResource(oneToOneB{})
	require.Nil(t, err)

	// create instance of oneToOneA
	res, err := resourceA.InsertInstance(app.DB(), &oneToOneA{Attr: "test"}).Result()
	require.Nil(t, err)
	a := res.(*oneToOneA)

	// create instance of oneToManyB with relation to a
	res, err = resourceB.InsertInstance(app.DB(), &oneToOneB{A: *a}).Result()
	require.Nil(t, err)
	b := res.(*oneToOneB)

	// ensure relation is properly set
	require.Equal(t, a.Id, b.A.Id)
	require.Equal(t, a.Attr, b.A.Attr)

	// fetch oneToManyA to update relations
	res, err = resourceA.SelectById(app.DB(), a.Id).Result()
	require.Nil(t, err)
	a = res.(*oneToOneA)

	// ensure relation is properly set
	require.Equal(t, b.Id, a.B.Id)
}
