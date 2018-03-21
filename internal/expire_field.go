package internal

import (
	"fmt"
	"github.com/go-pg/pg"
)

type expireField struct {
	*attrField
}

const expireNotificationChannel = "jargo_expire"

// expireTriggerQuery creates a trigger function
// that notifies the expire notification channel
// whenever a row was inserted, deleted or the
// value of its expire column changed.
const expireTriggerQuery = `
CREATE OR REPLACE FUNCTION jargo_expire_trigger_%s_func()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    PERFORM pg_notify('%s', json_build_object(
      'table', TG_TABLE_NAME,
      'id', NEW.id::text,
      'type', TG_OP,
      'now', NOW(),
      'expires', NEW."%s"
    )::text);
  ELSIF TG_OP = 'DELETE' THEN
    PERFORM pg_notify('%s', json_build_object(
      'table', TG_TABLE_NAME,
      'id', OLD.id::text,
      'type', TG_OP
    )::text);
  ELSIF TG_OP = 'UPDATE' THEN
    -- only send a notification if the
    -- value of the expire column actually changed
    IF NEW."%s" != OLD."%s" THEN
      PERFORM pg_notify('%s', json_build_object(
        'table', TG_TABLE_NAME,
        'id', NEW.id::text,
        'type', TG_OP,
        'now', NOW(),
        'expires', NEW."%s"
      )::text);
    END IF
  ELSE
    RAISE EXCEPTION 'Invalid Trigger Operation: %%', TG_OP;
  END IF
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS zzz_jargo_expire_trigger ON "%s";

CREATE TRIGGER zzz_jargo_expire_trigger
AFTER INSERT OR UPDATE OR DELETE ON "%s"
FOR EACH ROW EXECUTE PROCEDURE jargo_expire_trigger_%s_func();
`

func (f *expireField) afterCreateTable(db *pg.DB) error {
	_, err := db.Exec(fmt.Sprintf(expireTriggerQuery,
		f.schema.table,                      // function name
		expireNotificationChannel, f.column, // INSERT
		expireNotificationChannel,                               // DELETE
		f.column, f.column, expireNotificationChannel, f.column, // UPDATE
		f.schema.table,           // DROP TRIGGER
		f.schema.table, f.column, // CREATE TRIGGER
	))
	return err
}
