package internal

import "github.com/go-pg/pg"

type uuidIdField struct {
	*idField
}

const uuidExtensionQuery = `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`

func (f *uuidIdField) beforeCreateTable(db *pg.DB) error {
	_, err := db.Exec(uuidExtensionQuery)
	return err
}
