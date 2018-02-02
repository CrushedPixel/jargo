package internal

import "gopkg.in/go-playground/validator.v9"

// A SchemaInstance is an instance of a Schema
// holding values for each of the Schema fields.
type SchemaInstance struct {
	schema *Schema
	fields []schemaFieldInstance
}

// ToResourceModel creates a new Resource Model Instance
// from the fields of the schema instance.
func (i *SchemaInstance) ToResourceModel() interface{} {
	instance := i.schema.newResourceModelInstance()
	for _, f := range i.fields {
		f.applyToResourceModel(instance)
	}
	return instance.value.Interface()
}

// ToJsonapiModel creates a new Jsonapi Model Instance
// from the fields of the schema instance.
func (i *SchemaInstance) toJoinResourceModel() interface{} {
	instance := i.schema.newResourceModelInstance()
	for _, f := range i.fields {
		f.applyToJoinResourceModel(instance)
	}
	return instance.value.Interface()
}

// ToPGModel creates a new PG Model Instance
// from the fields of the schema instance.
func (i *SchemaInstance) ToJsonapiModel() interface{} {
	instance := i.schema.newJsonapiModelInstance()
	for _, f := range i.fields {
		f.applyToJsonapiModel(instance)
	}
	return instance.value.Interface()
}

func (i *SchemaInstance) toJoinJsonapiModel() interface{} {
	instance := i.schema.newJoinJsonapiModelInstance()
	for _, f := range i.fields {
		f.applyToJoinJsonapiModel(instance)
	}
	return instance.value.Interface()
}

func (i *SchemaInstance) ToPGModel() interface{} {
	instance := i.schema.newPGModelInstance()
	for _, f := range i.fields {
		f.applyToPGModel(instance)
	}
	return instance.value.Interface()
}

func (i *SchemaInstance) toJoinPGModel() interface{} {
	instance := i.schema.newJoinPGModelInstance()
	for _, f := range i.fields {
		f.applyToJoinPGModel(instance)
	}
	return instance.value.Interface()
}

// Validate validates a schema instance
// according to validator rules.
func (i *SchemaInstance) Validate(validate *validator.Validate) error {
	for _, f := range i.fields {
		err := f.validate(validate)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetRelationIds returns all direct (belongsTo) relations
// to other resource schemas.
func (i *SchemaInstance) GetRelationIds() map[*Schema][]int64 {
	m := make(map[*Schema][]int64)

	for _, f := range i.fields {
		if b, ok := f.(*belongsToFieldInstance); ok {
			if id, ok := b.relationId(); ok {
				m[b.relationSchema] = append(m[b.relationSchema], id)
			}
		}
	}

	return m
}
