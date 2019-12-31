// +build expire

package integration

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type expireAttribute struct {
	Id      int64
	Expires time.Time `jargo:",expire"`
}

// TestExpire tests the behaviour of expiration attributes.
func TestExpire(t *testing.T) {
	resource, err := app.RegisterResource(expireAttribute{})
	require.Nil(t, err)

	// insert resource with expiration set to 2s in the future
	original := &expireAttribute{
		Expires: time.Now().Add(2 * time.Second),
	}
	res, err := resource.InsertInstance(app.DB(), original).Result()
	require.Nil(t, err)

	// fetch resource to ensure it is still there
	res, err = resource.SelectById(app.DB(), res.(*expireAttribute).Id).Result()
	require.Nil(t, err)
	require.Equal(t, original.Expires.Unix(), res.(*expireAttribute).Expires.Unix())

	// sleep 3 seconds so resource must have timed out
	time.Sleep(3 * time.Second)

	// fetch resource again, expecting it to have timed out
	res, err = resource.SelectById(app.DB(), res.(*expireAttribute).Id).Result()
	require.Nil(t, err)
	require.Nil(t, res)

	// test update trigger functionality
	// insert resource with expiration set to 10h in the future
	original = &expireAttribute{
		Expires: time.Now().Add(10 * time.Hour),
	}
	res, err = resource.InsertInstance(app.DB(), original).Result()
	require.Nil(t, err)
	created := res.(*expireAttribute)

	// update resource, setting expiration time to 2s in the future
	created.Expires = time.Now().Add(2 * time.Second)
	res, err = resource.UpdateInstance(app.DB(), created).Result()
	require.Nil(t, err)

	// sleep 3 seconds so resource must have timed out
	time.Sleep(3 * time.Second)

	// fetch resource again, expecting it to have timed out
	res, err = resource.SelectById(app.DB(), created.Id).Result()
	require.Nil(t, err)
	require.Nil(t, res)
}

type multipleExpire struct {
	Id       int64
	ExpiresA time.Time `jargo:",expire"`
	ExpiresB time.Time `jargo:",expire"`
}

// TestMultipleExpire tests the behaviour of fields with multiple expire attributes.
func TestMultipleExpire(t *testing.T) {
	_, err := app.RegisterResource(multipleExpire{})
	require.EqualError(t, err, `error registering schema: "expire" option may not occur on multiple attributes`)
}

type defaultExpire struct {
	Id      int64
	Expires *time.Time `jargo:",expire,default:NOW() + INTERVAL '2 seconds'"`
}

// TestDefaultExpire tests the behaviour of expire fields with a default value.
func TestDefaultExpire(t *testing.T) {
	resource, err := app.RegisterResource(defaultExpire{})
	require.Nil(t, err)

	// insert resource with Expires field set to nil,
	// causing the default value to be inserted
	res, err := resource.InsertInstance(app.DB(), &defaultExpire{}).Result()
	require.Nil(t, err)
	created := res.(*defaultExpire)

	// fetch resource to ensure it is still there
	res, err = resource.SelectById(app.DB(), created.Id).Result()
	require.Nil(t, err)
	require.Equal(t, created.Id, res.(*defaultExpire).Id)

	// sleep 3 seconds so resource must have timed out
	time.Sleep(3 * time.Second)

	// fetch resource again, expecting it to have timed out
	res, err = resource.SelectById(app.DB(), created.Id).Result()
	require.Nil(t, err)
	require.Nil(t, res)
}
