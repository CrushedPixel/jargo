// +build integration

package integration

import (
	"testing"
	"flag"
	"github.com/go-pg/pg"
	"os"
	"crushedpixel.net/jargo"
)

const (
	clearQuery = `do
$$
declare
  l_stmt text;
  cnt integer;
begin
  select COUNT(*), 'drop ' || string_agg(format('%I.%I', schemaname, tablename), ',')
    into cnt, l_stmt
  from pg_tables
  where schemaname in ('public');
  if cnt > 0
  then
    execute l_stmt;
  end if;
end;
$$`
)

var (
	pgURL = flag.String("test_db", "postgres://jargo@localhost/jargo?sslmode=disable", "url for integration test database connection")

	app *jargo.Application
)

func TestMain(m *testing.M) {
	url, err := pg.ParseURL(*pgURL)
	if err != nil {
		panic(err)
	}

	// setup database
	db := pg.Connect(url)
	_, err = db.Exec(clearQuery)
	if err != nil {
		panic(err)
	}

	app = jargo.NewApplication(db)

	os.Exit(m.Run())
}
