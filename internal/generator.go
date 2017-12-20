package internal

import (
	"reflect"
	"fmt"
)

const (
	primaryFieldJsonapiName = "id"
	primaryFieldColumn      = "id"
)

// generates all fields that are always required to describe a resource via jsonapi,
// currently only the Id field
func generateStaticJsonapiFields(d *resourceDefinition) []reflect.StructField {
	for _, f := range d.fields {
		if f.typ != id {
			continue
		}

		tag := fmt.Sprintf(`jsonapi:"primary,%s"`, d.name)

		sf := reflect.StructField{
			Name: f.structField.Name,
			Type: f.structField.Type,
			Tag:  reflect.StructTag(tag),
		}

		return []reflect.StructField{sf}
	}

	panic("could not find id field")
}

// generates the struct fields required to describe an attribute field via jsonapi
func generateJsonapiFields(f *fieldDefinition) []reflect.StructField {
	sf := reflect.StructField{
		Name: f.structField.Name,
		Type: f.structField.Type,
	}

	tag := `jsonapi:"`
	switch f.typ {
	case attribute:
		tag += fmt.Sprintf(`attr,%s`, f.name)
	case has, belongsTo, many2many:
		tag += fmt.Sprintf(`relation,%s`, f.name)
	default:
		panic("can only generate jsonapi fields for member fields")
	}

	if f.jsonOmitempty {
		tag += `,omitempty`
	}

	tag += `"`

	sf.Tag = reflect.StructTag(tag)
	return []reflect.StructField{sf}
}

// generates all fields that are always required to describe a resource via go-pg,
// currently the TableName and Id field
func generateStaticPGFields(d *resourceDefinition) []reflect.StructField {
	empty := new(struct{})
	tableNameField := reflect.StructField{
		Name: "TableName",
		Type: reflect.TypeOf(*empty),
		Tag:  reflect.StructTag(fmt.Sprintf(`sql:"%s,alias:%s"`, d.table, d.alias)),
	}

	fields := []reflect.StructField{tableNameField}

	for _, f := range d.fields {
		if f.typ != id {
			continue
		}

		sf := reflect.StructField{
			Name: f.structField.Name,
			Type: f.structField.Type,
			Tag:  reflect.StructTag(fmt.Sprintf(`sql:"%s,pk"`, primaryFieldColumn)),
		}

		return append(fields, sf)
	}

	panic("could not find id field")
}

// generates the struct field required to describe an attribute field via go-pg.
// for belongsTo relation fields, generates the id field for the foreign value.
func generatePGFields(f *fieldDefinition, r *Registry) []reflect.StructField {
	fields := make([]reflect.StructField, 0)

	sf := reflect.StructField{
		Name: f.structField.Name,
		Type: f.structField.Type,
	}

	tag := ""
	switch f.typ {
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
		// get id field of relation from registry to determine id type
		res, err := r.getResource(getStructType(f.structField.Type))
		if err != nil {
			panic(err)
		}

		var idFieldType reflect.Type
		for _, f0 := range res.definition.fields {
			if f0.typ == id {
				idFieldType = f0.structField.Type
			}
		}
		if idFieldType == nil {
			panic("could not get id type of relation")
		}

		idField := reflect.StructField{
			Name: fmt.Sprintf("%sId", f.structField.Name),
			Type: idFieldType,
		}

		fields = append(fields, idField)
	case many2many:
		tag = fmt.Sprintf(`pg:",many2many:%s"`, f.pgJoinTable)
	default:
		panic("can only generate pg fields for member fields")
	}

	sf.Tag = reflect.StructTag(tag)
	return append(fields, sf)
}
