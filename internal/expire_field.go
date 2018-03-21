package internal

import (
	"fmt"
	"github.com/go-pg/pg"
)

// ExpireNotificationChannelName returns the expire
// notification channel name for a given database table.
func ExpireNotificationChannelName(table string) string {
	return fmt.Sprintf("jargo_expire_%s", table)
}

type expireField struct {
	*attrField
}

// expireTriggerQuery creates a trigger function
// that notifies the expire notification channel
// whenever a row was inserted, deleted or the
// value of its expire column changed.
const expireTriggerQuery = `
CREATE OR REPLACE FUNCTION jargo_expire_trigger_%s_func()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' OR (TG_OP = 'UPDATE' AND NEW."%s" != OLD."%s") THEN
    SELECT EXTRACT(EPOCH FROM (NEW."%s" - NOW())) AS "interval" INTO interval;
    PERFORM pg_notify('%s', json_build_object(
      'type', TG_OP,
      'interval', interval."interval"
    )::text);
  END IF
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS zzz_jargo_expire_trigger ON "%s";

CREATE TRIGGER zzz_jargo_expire_trigger
AFTER INSERT OR UPDATE ON "%s"
FOR EACH ROW EXECUTE PROCEDURE jargo_expire_trigger_%s_func();
`

func (f *expireField) afterCreateTable(db *pg.DB) error {
	_, err := db.Exec(fmt.Sprintf(expireTriggerQuery,
		f.schema.table,           // function name
		f.column, f.schema.table, // IF EXPIRED DELETE
		ExpireNotificationChannelName(f.schema.table), f.column, // INSERT
		ExpireNotificationChannelName(f.schema.table), // DELETE
		f.column, f.column, ExpireNotificationChannelName(f.schema.table), f.column, // UPDATE
		f.schema.table,           // DROP TRIGGER
		f.schema.table, f.column, // CREATE TRIGGER
	))
	return err
}
