package resource

import (
	"reflect"
	"fmt"
	"github.com/pkg/errors"
)

// TODO: allow for omitempty
func generateJsonapiStructType(definition *resourceDefinition) reflect.Type {
	var fields []reflect.StructField

	for _, f := range definition.fields {
		sf := reflect.StructField{
			Name: f.structField.Name,
			Type: f.structField.Type,
		}

		var tag string
		switch f.typ {
		case id:
			tag = fmt.Sprintf(`jsonapi:"primary,%s"`, definition.name)
		case attribute:
			tag = fmt.Sprintf(`jsonapi:"attr,%s"`, f.name)
		case has, belongsTo, many2many:
			tag = fmt.Sprintf(`jsonapi:"relation,%s"`, f.name)
		}

		sf.Tag = reflect.StructTag(tag)
		fields = append(fields, sf)
	}

	return reflect.StructOf(fields)
}

func generatePGModel(r *Registry, definition *resourceDefinition) reflect.Type {
	var fields []reflect.StructField
	// specify table name and alias
	str := new(struct{})
	tableNameField := reflect.StructField{
		Name: "TableName",
		Type: reflect.TypeOf(*str),
		Tag:  reflect.StructTag(fmt.Sprintf(`sql:"%s,alias:%s"`, definition.table, definition.alias)),
	}
	fields = append(fields, tableNameField)

	for _, f := range definition.fields {
		sf := reflect.StructField{
			Name: f.structField.Name,
			Type: f.structField.Type,
		}

		tag := ""
		switch f.typ {
		case id:
			tag = `sql:",pk"`
		case attribute:
			tag = fmt.Sprintf(`sql:"%s`, f.column)

			if f.sqlNotnull {
				tag += ",notnull"
			}
			if f.sqlUnique {
				tag += ",unique"
			}
			if f.sqlDefault != "" {
				tag += fmt.Sprintf(",default:%s", f.sqlDefault)
			}

			tag += `"`
		case has:
			if f.pgFk != "" {
				tag = fmt.Sprintf(`pg:",fk:%s"`, f.pgFk)
			}
		case belongsTo:
			// generate id field holding id of relation field
			res, err := r.getResource(getStructType(f.structField.Type))
			if err != nil {
				panic(err)
			}

			// get id field of relation and get its type
			var idFieldType reflect.Type
			for _, f0 := range res.definition.fields {
				if f0.typ == id {
					idFieldType = f0.structField.Type
				}
			}
			if idFieldType == nil {
				panic(errors.New("could not get id type of relation"))
			}

			idField := reflect.StructField{
				Name: fmt.Sprintf("%sId", f.structField.Name),
				Type: idFieldType,
			}

			fields = append(fields, idField)
		case many2many:
			tag = fmt.Sprintf(`pg:",many2many:%s"`, f.pgJoinTable)
		}

		sf.Tag = reflect.StructTag(tag)
		fields = append(fields, sf)
	}

	return reflect.StructOf(fields)
}
