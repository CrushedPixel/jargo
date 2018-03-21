package jargo

import (
	"context"
	"fmt"
	"github.com/crushedpixel/jargo/internal"
	"github.com/go-pg/pg"
	"github.com/json-iterator/go"
	"log"
	"sync"
	"time"
)

// notificationPayload is a struct representation
// of the json payload sent over expire notification channels.
type expireNotificationPayload struct {
	// Interval is the amount of seconds
	// until the next expiration is reached
	Interval float64 `json:"interval"`
}

// resourceExpirer is responsible for expiring
// a Resource's records.
type resourceExpirer struct {
	app *Application
	// resource table name for queries
	table string
	// resource alias for queries
	alias string
	// SQL-safe expire column name for queries
	column string

	// expirationChannel is the channel
	// to send the next expiration times into.
	expirationChannel chan time.Time

	// lock protects all variables below.
	lock *sync.Mutex
	// nextExpiration is the time at which the next expiration happens.
	nextExpiration time.Time
	// cancel is the CancelFunc for the task scheduling the next expiration.
	cancel context.CancelFunc
}

// newResourceExpirer creates and starts a new resourceExpirer.
func newResourceExpirer(ctx context.Context, app *Application, resource *Resource) *resourceExpirer {
	e := &resourceExpirer{
		app:    app,
		table:  resource.schema.Table(),
		alias:  resource.schema.Alias(),
		column: escapePGColumn(resource.schema.ExpireField().PGFilterColumn()),

		expirationChannel: make(chan time.Time, 0),

		lock: new(sync.Mutex),
	}

	// expiration channel
	expireNotificationChannel := app.DB().Listen(
		internal.ExpireNotificationChannelName(e.table),
	).Channel()

	// start expiration loop
	go e.expirationLoop(ctx, expireNotificationChannel)

	// fetch initial expiration time from database
	go e.fetchNextExpiration(ctx)

	return e
}

func (e *resourceExpirer) expirationLoop(ctx context.Context, notificationChannel <-chan *pg.Notification) {
	for {
		select {
		case nextExpiration := <-e.expirationChannel:
			e.scheduleNextExpiration(ctx, nextExpiration)
		case notification := <-notificationChannel:
			// received a notification from the expire database trigger
			payload := &expireNotificationPayload{}
			if err := jsoniter.Unmarshal([]byte(notification.Payload), payload); err != nil {
				panic(err)
			}
			ne := targetTime(payload.Interval)
			e.lock.Lock()
			// if next expiration time as reported by the trigger
			// is smaller than the known next expiration time,
			// send it into the expiration channel.
			if e.nextExpiration.IsZero() || ne.Before(e.nextExpiration) {
				go func() { e.expirationChannel <- ne }()
			}
			e.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// scheduleNextExpiration cancels the current expiration task
// and schedules the next expiration to happen at expirationTime.
//
// If expirationTime is the zero value of time.Time, no expiration is scheduled.
func (e *resourceExpirer) scheduleNextExpiration(ctx context.Context, expirationTime time.Time) {
	e.lock.Lock()

	// set next expiration time
	e.nextExpiration = expirationTime

	// cancel ongoing expiration task, if any
	if e.cancel != nil {
		e.cancel()
		e.cancel = nil
	}

	if !expirationTime.IsZero() {
		// create child context of resourceExpirer's context
		childCtx, cancel := context.WithCancel(ctx)
		e.cancel = cancel
		go e.expirationTask(ctx, childCtx, expirationTime)
	}

	e.lock.Unlock()
}

// expirationTask expires all expired records once expirationTime is reached.
func (e *resourceExpirer) expirationTask(parentCtx, ctx context.Context, expirationTime time.Time) {
	select {
	case <-time.After(expirationTime.Sub(time.Now())):
		query := fmt.Sprintf(`DELETE FROM "%s" AS "%s" WHERE %s <= NOW()`,
			e.table, e.alias, e.column)
		if _, err := e.app.DB().Exec(query); err != nil {
			log.Printf("Error expiring records: %s\n", err.Error())
		}

		// fetch next expiration time
		e.fetchNextExpiration(parentCtx)
	case <-ctx.Done():
	}
}

// fetchNextExpiration fetches the time at which the next record expires.
// Returns time.Time's zero value if there is no record to expire.
func (e *resourceExpirer) fetchNextExpiration(ctx context.Context) {
	mdl := &struct {
		// Interval is the amount of seconds
		// until the next expiration is reached
		Interval float64
	}{}

	var nextExpiration time.Time
	// fetch amount of seconds until next expiration time is reached
	query := fmt.Sprintf(`SELECT EXTRACT(EPOCH FROM (%s)) AS "interval" FROM "%s" AS "%s" ORDER BY %s ASC LIMIT 1`,
		e.column, e.table, e.alias, e.column)
	if _, err := e.app.DB().QueryOne(mdl, query); err == nil {
		nextExpiration = targetTime(mdl.Interval)
	} else {
		if err != pg.ErrNoRows {
			log.Printf("Error fetching next expiration time: %s\n", err.Error())
		}
	}

	// send expiration time into channel
	// unless context is done
	select {
	case <-ctx.Done():
	default:
		e.expirationChannel <- nextExpiration
	}
}

// targetTime returns the time that will be in sec seconds.
func targetTime(sec float64) time.Time {
	return time.Now().Add(time.Duration(float64(time.Second) * sec))
}
