// +build integration

package integration

import (
	"context"
	"github.com/crushedpixel/jargo"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type realtimeTestA struct {
	Id int64
	Bs []realtimeTestB `jargo:",has:A"`
}

type realtimeTestB struct {
	Id int64
	A  realtimeTestA `jargo:",belongsTo"`
}

func TestRealtime(t *testing.T) {
	resourceA, err := app.RegisterResource(realtimeTestA{})
	require.Nil(t, err)

	resourceB, err := app.RegisterResource(realtimeTestB{})
	require.Nil(t, err)

	realtime := jargo.NewRealtime(app, "/realtime")

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	go func(ctx context.Context) {
		realtime.Run(ctx)
	}(ctx)

	time.Sleep(1 * time.Second)

	// triggers INSERT of resourceA
	res, err := resourceA.InsertInstance(app.DB(), &realtimeTestA{Id: 100}).Result()
	require.Nil(t, err)
	a := res.(*realtimeTestA)

	// triggers UPDATE of resourceA
	res, err = resourceB.InsertInstance(app.DB(), &realtimeTestB{A: *a}).Result()
	require.Nil(t, err)

	// triggers DELETE of resourceB and UPDATE of resourceA
	_, err = resourceB.DeleteById(app.DB(), res.(*realtimeTestB).Id).Result()
	require.Nil(t, err)

	time.Sleep(1 * time.Second)
}
