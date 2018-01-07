// +build integration

package integration

import (
	"testing"
	"flag"
	"github.com/go-pg/pg"
	"os"
	"crushedpixel/jargo"
	"log"
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
	pgURL = flag.String("test_db", "postgres://jargo@localhost/jargo?sslmode=disable", "url for integration test database connection")

	app *jargo.Application
)

func TestMain(m *testing.M) {
	flag.Parse()

	url, err := pg.ParseURL(*pgURL)
	if err != nil {
		panic(err)
	}

	// setup database
	db := pg.Connect(url)

	// log database queries
	db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
		query, err := event.FormattedQuery()
		if err != nil {
			panic(err)
		}

		log.Printf("%s %s", time.Since(event.StartTime), query)
	})

	_, err = db.Exec(clearQuery)
	if err != nil {
		panic(err)
	}

	app = jargo.NewApplication(db)

	os.Exit(m.Run())
}
