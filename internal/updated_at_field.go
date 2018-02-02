package internal

import (
	"fmt"
	"github.com/go-pg/pg"
)

type updatedAtField struct {
	*attrField
}

const updatedAtTriggerQuery = `
CREATE OR REPLACE FUNCTION jargo_updated_at_trigger_%s_func()
RETURNS TRIGGER AS
$$
BEGIN
  NEW."%s" := NOW();
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS jargo_updated_at_trigger_%s ON "%s";

CREATE TRIGGER jargo_updated_at_trigger_%s
BEFORE UPDATE ON "%s"
FOR EACH ROW EXECUTE PROCEDURE jargo_updated_at_trigger_%s_func();
`

func (f *updatedAtField) afterCreateTable(db *pg.DB) error {
	_, err := db.Exec(fmt.Sprintf(updatedAtTriggerQuery,
		f.column, f.column,                 // CREATE FUNCTION statement
		f.column, f.schema.table,           // DROP TRIGGER statement
		f.column, f.schema.table, f.column, // CREATE TRIGGER statement
	))
	return err
}
