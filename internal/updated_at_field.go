package internal

import (
	"github.com/go-pg/pg"
	"fmt"
)

type updatedAtField struct {
	*attrField
}

const triggerFunctionQuery = `
CREATE OR REPLACE FUNCTION jargo_updated_at_trigger_%s_func()
RETURNS TRIGGER AS
$$
BEGIN
  NEW."%s" := NOW();
  RETURN NEW;
END
$$ LANGUAGE plpgsql;
`

const dropTriggerQuery = `
DROP TRIGGER IF EXISTS jargo_updated_at_trigger_%s ON "%s"
`

const createTriggerQueryFormat = `
CREATE TRIGGER jargo_updated_at_trigger_%s
BEFORE UPDATE ON "%s"
FOR EACH ROW EXECUTE PROCEDURE jargo_updated_at_trigger_%s_func();
`

func (f *updatedAtField) afterCreateTable(db *pg.DB) error {
	return db.RunInTransaction(func(tx *pg.Tx) error {
		_, err := tx.Exec(fmt.Sprintf(triggerFunctionQuery, f.column, f.column))
		if err != nil {
			return err
		}

		_, err = tx.Exec(fmt.Sprintf(dropTriggerQuery, f.column, f.schema.table))
		if err != nil {
			return err
		}

		_, err = tx.Exec(fmt.Sprintf(createTriggerQueryFormat, f.column, f.schema.table, f.column))
		if err != nil {
			return err
		}

		return nil
	})
}
