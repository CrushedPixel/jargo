// +build integration

package integration

import (
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

type oneToManyA struct {
	Id   string
	Attr string
	Bs   []oneToManyB `jargo:",has:A"`
}

type oneToManyB struct {
	Id int64
	A  oneToManyA `jargo:",belongsTo"`
}

// TestOneToManyRelations tests the behaviour of one-to-many relationships.
func TestOneToManyRelations(t *testing.T) {
	resourceA, err := app.RegisterResource(oneToManyA{})
	require.Nil(t, err)

	resourceB, err := app.RegisterResource(oneToManyB{})
	require.Nil(t, err)

	// create instance of oneToManyA
	res, err := resourceA.InsertInstance(app.DB(), &oneToManyA{Id: "parent", Attr: "test"}).Result()
	require.Nil(t, err)
	a := res.(*oneToManyA)

	// ensure instance properly encodes to json
	json, err := resourceA.ResponseAllFields(a).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"one-to-many-as","id":"parent","attributes":{"attr":"test"},"relationships":{"bs":{"data":[]}}}}`,
		json)

	// create instance of oneToManyB with relation to a
	res, err = resourceB.InsertInstance(app.DB(), &oneToManyB{A: *a}).Result()
	require.Nil(t, err)
	b := res.(*oneToManyB)

	// ensure relation is properly set
	require.Equal(t, a.Id, b.A.Id)
	require.Equal(t, a.Attr, b.A.Attr)

	// ensure relation properly encodes to json
	json, err = resourceB.ResponseAllFields(b).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"one-to-many-bs","id":"1","relationships":{"a":{"data":{"type":"one-to-many-as","id":"parent"}}}}}`,
		json)

	// fetch oneToManyA to update relations
	res, err = resourceA.SelectById(app.DB(), a.Id).Result()
	require.Nil(t, err)
	a = res.(*oneToManyA)

	// ensure relation is properly set
	require.Equal(t, b.Id, a.Bs[0].Id)

	// ensure relation properly encodes to json
	json, err = resourceA.ResponseAllFields(a).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"one-to-many-as","id":"parent","attributes":{"attr":"test"},"relationships":{"bs":{"data":[{"type":"one-to-many-bs","id":"1"}]}}}}`,
		json)
}

type oneToManyNullableA struct {
	Id int64
	B  []oneToManyNullableB `jargo:",has:A"`
}

type oneToManyNullableB struct {
	Id int64
	A  *oneToManyNullableA `jargo:",belongsTo"`
}

// TestOneToManyNullableRelations tests the behaviour
// of one-to-many nullable relationships.
func TestOneToManyNullableRelations(t *testing.T) {
	resourceA, err := app.RegisterResource(oneToManyNullableA{})
	require.Nil(t, err)

	resourceB, err := app.RegisterResource(oneToManyNullableB{})
	require.Nil(t, err)

	// create instance of oneToManyNullableB with the relation set to null
	res, err := resourceB.InsertInstance(app.DB(), &oneToManyNullableB{}).Result()
	require.Nil(t, err)
	b := res.(*oneToManyNullableB)

	// ensure instance properly encodes to json
	json, err := resourceB.ResponseAllFields(b).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"one-to-many-nullable-bs","id":"1","relationships":{"a":{"data":null}}}}`,
		json)

	// create instance of oneToManyNullableA
	res, err = resourceA.InsertInstance(app.DB(), &oneToManyNullableA{}).Result()
	require.Nil(t, err)
	a := res.(*oneToManyNullableA)

	// update b to reference a
	b.A = a
	res, err = resourceB.UpdateInstance(app.DB(), b).Result()
	require.Nil(t, err)
	b = res.(*oneToManyNullableB)

	// ensure instance properly encodes to json
	json, err = resourceB.ResponseAllFields(b).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"one-to-many-nullable-bs","id":"1","relationships":{"a":{"data":{"type":"one-to-many-nullable-as","id":"1"}}}}}`,
		json)
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

// TestOneToOneRelations tests the behaviour of one-to-one relationships.
func TestOneToOneRelations(t *testing.T) {
	resourceA, err := app.RegisterResource(oneToOneA{})
	require.Nil(t, err)

	resourceB, err := app.RegisterResource(oneToOneB{})
	require.Nil(t, err)

	// create instance of oneToOneA
	res, err := resourceA.InsertInstance(app.DB(), &oneToOneA{Attr: "test"}).Result()
	require.Nil(t, err)
	a := res.(*oneToOneA)

	// ensure instance properly encodes to json
	json, err := resourceA.ResponseAllFields(a).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"one-to-one-as","id":"1","attributes":{"attr":"test"},"relationships":{"b":{"data":null}}}}`,
		json)

	// create instance of oneToManyB with relation to a
	res, err = resourceB.InsertInstance(app.DB(), &oneToOneB{A: *a}).Result()
	require.Nil(t, err)
	b := res.(*oneToOneB)

	// ensure relation is properly set
	require.Equal(t, a.Id, b.A.Id)
	require.Equal(t, a.Attr, b.A.Attr)

	// ensure relation properly encodes to json
	json, err = resourceB.ResponseAllFields(b).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"one-to-one-bs","id":"1","relationships":{"a":{"data":{"type":"one-to-one-as","id":"1"}}}}}`,
		json)

	// fetch oneToManyA to update relations
	res, err = resourceA.SelectById(app.DB(), a.Id).Result()
	require.Nil(t, err)
	a = res.(*oneToOneA)

	// ensure relation is properly set
	require.Equal(t, b.Id, a.B.Id)

	// ensure relation properly encodes to json
	json, err = resourceA.ResponseAllFields(a).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"one-to-one-as","id":"1","attributes":{"attr":"test"},"relationships":{"b":{"data":{"type":"one-to-one-bs","id":"1"}}}}}`,
		json)
}

type oneToSelf struct {
	Id       int64
	Parent   *oneToSelf  `jargo:",belongsTo"`
	Children []oneToSelf `jargo:",has:Parent"`
}

// TestOneToSelfRelations tests the behaviour of relationships with itself.
func TestOneToSelfRelations(t *testing.T) {
	resource, err := app.RegisterResource(oneToSelf{})
	require.Nil(t, err)

	// create instance in database
	res, err := resource.InsertInstance(app.DB(), &oneToSelf{}).Result()
	require.Nil(t, err)
	parent := res.(*oneToSelf)

	// create multiple children in database
	childCount := 5
	for i := 0; i < childCount; i++ {
		res, err := resource.InsertInstance(app.DB(), &oneToSelf{Parent: parent}).Result()
		require.Nil(t, err)

		child := res.(*oneToSelf)
		require.Equal(t, parent.Id, child.Parent.Id)

		// ensure relation properly encodes to json
		json, err := resource.ResponseAllFields(child).Payload()
		require.Nil(t, err)
		require.Equal(t,
			fmt.Sprintf(`{"data":{"type":"one-to-selves","id":"%d","relationships":{"children":{"data":[]},"parent":{"data":{"type":"one-to-selves","id":"1"}}}}}`, i+2),
			json)
	}

	// fetch parent from database to ensure children are set
	res, err = resource.SelectById(app.DB(), parent.Id).Result()
	require.Nil(t, err)

	parent = res.(*oneToSelf)
	require.Len(t, parent.Children, childCount)

	// ensure relation properly encodes to json
	json, err := resource.ResponseAllFields(parent).Payload()
	require.Nil(t, err)
	require.Equal(t,
		`{"data":{"type":"one-to-selves","id":"1","relationships":{"children":{"data":[{"type":"one-to-selves","id":"2"},{"type":"one-to-selves","id":"3"},{"type":"one-to-selves","id":"4"},{"type":"one-to-selves","id":"5"},{"type":"one-to-selves","id":"6"}]},"parent":{"data":null}}}}`,
		json)
}

type belongsToA struct {
	Id int64
	B  belongsToB `jargo:",belongsTo"`
}

type belongsToB struct {
	Id int64
}

// TestBelongsToRelations tests the behaviour of belongsTo relations.
func TestBelongsToRelations(t *testing.T) {
	_, err := app.RegisterResource(belongsToA{})
	require.Nil(t, err)
}

type hasA struct {
	Id uuid.UUID
	B  []hasB `jargo:",has:A"`
}

type hasB struct {
	Id uuid.UUID
	A  hasA  `jargo:",belongsTo"`
	C  *hasC `jargo:",belongsTo"`
}

type hasC struct {
	Id uuid.UUID
}

// TestUUIDRelations tests the behaviour of relations with UUID id fields.
func TestUUIDRelations(t *testing.T) {
	resourceA, err := app.RegisterResource(hasA{})
	require.Nil(t, err)

	resourceB, err := app.RegisterResource(hasB{})
	require.Nil(t, err)

	// insert resource instance
	res, err := resourceA.InsertInstance(app.DB(), &hasA{}).Result()
	require.Nil(t, err)
	a := res.(*hasA)

	// fetch resource instance
	res, err = resourceA.SelectById(app.DB(), a.Id).Result()
	require.Nil(t, err)
	a = res.(*hasA)

	// insert resource B with relation to Resource A but nil relation to Resource C
	res, err = resourceB.InsertInstance(app.DB(), &hasB{A: *a}).Result()
	require.Nil(t, err)
}
