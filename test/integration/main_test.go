// +build integration

package integration

import (
	"flag"
	"github.com/crushedpixel/jargo"
	"github.com/go-pg/pg"
	"log"
	"os"
	"testing"
	"time"
)

const clearQuery = `
DO
$$
DECLARE
  l_stmt text;
  cnt integer;
BEGIN
  SELECT COUNT(*), 'DROP TABLE ' || string_agg(format('"%I"."%I"', schemaname, tablename), ',')
    INTO cnt, l_stmt
  FROM pg_tables
  WHERE schemaname IN ('public');
  IF cnt > 0
  THEN
    EXECUTE l_stmt;
  END IF;
END;
$$
`

var (
	pgURLFlag = flag.String("test_db", "postgres://jargo@localhost/jargo?sslmode=disable", "url for integration test database connection")
	debugFlag = flag.Bool("debug", false, "enable debug output")

	app *jargo.Application

	dummyResource *jargo.Resource
	// dummyInstance is an instance
	// of dummy to use for testing
	dummyInstance *dummy
)

func TestMain(m *testing.M) {
	flag.Parse()

	url, err := pg.ParseURL(*pgURLFlag)
	if err != nil {
		panic(err)
	}

	// setup database
	db := pg.Connect(url)

	if *debugFlag {
		// log database queries
		db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
			query, err := event.FormattedQuery()
			if err != nil {
				panic(err)
			}

			log.Printf("%s %s", time.Since(event.StartTime), query)
		})
	}

	// drop entire database
	_, err = db.Exec(clearQuery)
	if err != nil {
		panic(err)
	}

	app = jargo.NewApplication(jargo.Options{
		DB: db,
	})

	dummyResource = app.MustRegisterResource(dummy{})
	res, err := dummyResource.InsertInstance(app.DB(), &dummy{}).Result()
	if err != nil {
		panic(err)
	}
	dummyInstance = res.(*dummy)

	os.Exit(m.Run())
}

// dummy is an empty type for testing purposes
// in relations
type dummy struct {
	Id int64
}
