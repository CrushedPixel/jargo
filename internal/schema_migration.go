package internal

import (
	"fmt"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"reflect"
)

const migrationTableSuffix = "__migration"

type pgColumnInfo struct {
	TableName struct{} `sql:"information_schema.columns"`

	ColumnName    string
	IsNullable    bool
	DataType      string
	ColumnDefault string
}

func (s *Schema) performMigration(db *pg.DB) error {
	// to be able to check if the table columns of the model
	// are different than the existing table's columns,
	// we need to create a temporary table for the model
	// and compare its columns with the existing table's columns.

	// to create a table for the model under a different name,
	// we create a version of the pg model type struct
	// with the "TableName" field modified to contain
	// the temporary table name in the sql struct tag.
	tmpTableName := s.table + migrationTableSuffix

	var tmpFields []reflect.StructField
	for i := 0; i < s.pgModelType.NumField(); i++ {
		field := s.pgModelType.Field(i)
		if field.Name == pgTableNameFieldName {
			field.Tag = reflect.StructTag(fmt.Sprintf(`sql:"\"%s\""`, tmpTableName))
		}

		tmpFields = append(tmpFields, field)
	}
	tmpPgModelType := reflect.StructOf(tmpFields)

	// create the temporary table. this can't be done in a transaction,
	// as the information_schema is not updated until committing.
	if err := db.CreateTable(reflect.New(tmpPgModelType).Interface(),
		&orm.CreateTableOptions{Temp: true}); err != nil {
		return err
	}

	// get column information for the new table
	var newColumns []pgColumnInfo
	if err := db.Model(&newColumns).
		Where("table_name = ?", tmpTableName).
		Select(); err != nil {
		return err
	}

	// get column information for the existing table
	var oldColumns []pgColumnInfo
	if err := db.Model(&oldColumns).
		Where("table_name = ?", s.table).
		Select(); err != nil {
		return err
	}

	// compare the new table's columns with the old table's columns
	if reflect.DeepEqual(oldColumns, newColumns) {
		// the old and the new table are equal.
		// delete the temporary table.
		_, err := db.Exec(fmt.Sprintf(`DROP TABLE "%s"`, tmpTableName))
		return err
	}

	// the new table has a different schema.
	if err := db.RunInTransaction(func(tx *pg.Tx) error {
		// try to copy all data from the old table's columns into the new table
		var columnNames string
		for i, column := range oldColumns {
			if i > 0 {
				columnNames += ", "
			}
			columnNames += fmt.Sprintf(`"%s"`, column.ColumnName)
		}

		if _, err := tx.Exec(fmt.Sprintf(`INSERT INTO "%s" (%s) SELECT * FROM "%s"`, tmpTableName, columnNames, s.table)); err != nil {
			return err
		}

		if _, err := tx.Exec(fmt.Sprintf(`DROP TABLE "%s"`, s.table)); err != nil {
			return err
		}

		if _, err := tx.Exec(fmt.Sprintf(`ALTER TABLE "%s" RENAME TO "%s"`, tmpTableName, s.table)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		// migration failed.
		// drop the temporary table
		db.Exec(fmt.Sprintf(`DROP TABLE "%s"`, tmpTableName))

		return fmt.Errorf("table migration failed. you have to manually perform migration: %s", err.Error())
	}

	return nil
}
