package internal

import (
	"errors"
	"fmt"
	"github.com/go-pg/pg"
	"github.com/google/jsonapi"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"reflect"
)

// resource model -> resource Schema
// resource Schema instance <-> jsonapi model, pg model, resource model instances

var (
	errInvalidResourceInstance    = errors.New("instance must be pointer to resource model")
	errInvalidResourceCollection  = errors.New("collection must be slice of pointers to resource model")
	errInvalidJsonapiInstance     = errors.New("instance must be pointer to jsonapi model")
	errInvalidJsonapiCollection   = errors.New("collection must be slice of pointers to jsonapi model")
	errInvalidJoinJsonapiInstance = errors.New("instance must be pointer to join jsonapi model")
	errInvalidPGInstance          = errors.New("instance must be pointer to pg model")
	errInvalidPGCollection        = errors.New("collection must be slice of pointers to pg model")
	errInvalidJoinPGInstance      = errors.New("instance must be pointer to join pg model")
)

const realtimeTriggerQuery = `
DROP TRIGGER IF EXISTS jargo_%s_notify ON "%s";
CREATE TRIGGER jargo_%s_notify AFTER INSERT OR UPDATE OR DELETE ON "%s" FOR EACH ROW EXECUTE PROCEDURE %s();
`

type Schema struct {
	name  string // jsonapi member name
	table string // sql table name
	alias string // sql table alias

	fields []SchemaField

	resourceModelType reflect.Type
	jsonapiModelType  reflect.Type
	pgModelType       reflect.Type

	// model types to be referenced in relations
	// from other jsonapi/pg models,
	// avoiding infinite recursion
	joinJsonapiModelType reflect.Type
	joinPGModelType      reflect.Type
}

// JSONAPI returns the Schema's JSON API member name.
func (s *Schema) JSONAPIName() string {
	return s.name
}

// Table returns the Schema's table name in the database.
func (s *Schema) Table() string {
	return s.table
}

// Alias returns the Schema's alias used in queries.
func (s *Schema) Alias() string {
	return s.alias
}

// Fields returns all of Schema's fields.
func (s *Schema) Fields() []SchemaField {
	return s.fields
}

// IdField returns the Schema's id field.
func (s *Schema) IdField() SchemaField {
	for _, f := range s.fields {
		if f.JSONAPIName() == IdFieldJsonapiName {
			return f
		}
	}
	panic("could not find id field")
}

// ExpireField returns the Schema's expire field.
// Returns nil if the Schema has no expire field.
func (s *Schema) ExpireField() SchemaField {
	for _, f := range s.fields {
		if _, ok := f.(*expireField); ok {
			return f
		}
	}
	return nil
}

// CreateTable creates the database table
// for this Schema if it doesn't exist yet.
// It also implements primitive migration efforts,
// creating columns that don't yet exist.
func (s *Schema) CreateTable(db *pg.DB) error {
	// call beforeCreateTable hooks on fields
	for _, f := range s.Fields() {
		if hook, ok := f.(beforeCreateTableHook); ok {
			if err := hook.beforeCreateTable(db); err != nil {
				return err
			}
		}
	}

	// create table
	if err := db.CreateTable(s.NewPGModelInstance(), nil); err != nil {
		if pgErr, ok := err.(pg.Error); ok && pgErr.Field('C') == "42P07" {
			// the table already exists - perform migration
			if err := s.performMigration(db); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// call afterCreateTable hooks on fields
	for _, f := range s.Fields() {
		if hook, ok := f.(afterCreateTableHook); ok {
			if err := hook.afterCreateTable(db); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Schema) CreateRealtimeTriggers(db *pg.DB, functionName string) error {
	_, err := db.Exec(fmt.Sprintf(realtimeTriggerQuery,
		s.table, s.table, s.table, s.table, functionName,
	))
	return err
}

func (s *Schema) IsResourceModelCollection(data interface{}) bool {
	typ := reflect.ValueOf(data).Type()
	collection := false
	if typ.Kind() == reflect.Slice {
		collection = true
		typ = typ.Elem()
	}

	if typ != reflect.PtrTo(s.resourceModelType) {
		if collection {
			panic(errInvalidResourceCollection)
		} else {
			panic(errInvalidResourceInstance)
		}
	}

	return collection
}

func (s *Schema) IsJsonapiModelCollection(data interface{}) bool {
	typ := reflect.ValueOf(data).Type()
	collection := false
	if typ.Kind() == reflect.Slice {
		collection = true
		typ = typ.Elem()
	}

	if typ != reflect.PtrTo(s.jsonapiModelType) {
		if collection {
			panic(errInvalidJsonapiCollection)
		} else {
			panic(errInvalidJsonapiInstance)
		}
	}

	return collection
}

func (s *Schema) IsPGModelCollection(data interface{}) bool {
	typ := reflect.ValueOf(data).Type()
	collection := false
	if typ.Kind() == reflect.Slice {
		collection = true
		typ = typ.Elem()
	}

	if typ != reflect.PtrTo(s.pgModelType) {
		if collection {
			panic(errInvalidPGCollection)
		} else {
			panic(errInvalidPGInstance)
		}
	}

	return collection
}

func (s *Schema) NewResourceModelInstance() interface{} {
	return reflect.New(s.resourceModelType).Interface()
}

func (s *Schema) NewResourceModelCollection(entries ...interface{}) interface{} {
	l := len(entries)
	val := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(s.resourceModelType)), l, l)
	for i := 0; i < l; i++ {
		val.Index(i).Set(reflect.ValueOf(entries[i]))
	}
	return val.Interface()
}

func (s *Schema) NewJsonapiModelInstance() interface{} {
	return reflect.New(s.jsonapiModelType).Interface()
}

func (s *Schema) NewPGModelInstance() interface{} {
	return reflect.New(s.pgModelType).Interface()
}

func (s *Schema) NewPGModelCollection(entries ...interface{}) interface{} {
	l := len(entries)
	val := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(s.pgModelType)), l, l)
	for i := 0; i < l; i++ {
		val.Index(i).Set(reflect.ValueOf(entries[i]))
	}
	return val.Interface()
}

func (s *Schema) ParseResourceModelCollection(instance interface{}) []*SchemaInstance {
	collection := s.IsResourceModelCollection(instance)
	if !collection {
		panic(errInvalidResourceCollection)
	}

	v := reflect.ValueOf(instance)
	if v.IsNil() {
		return nil
	}

	var schemaInstances []*SchemaInstance
	for i := 0; i < v.Len(); i++ {
		child := v.Index(i)
		schemaInstance := s.ParseResourceModel(child.Interface())
		if schemaInstance != nil {
			schemaInstances = append(schemaInstances, schemaInstance)
		}
	}

	return schemaInstances
}

func (s *Schema) ParseResourceModel(instance interface{}) *SchemaInstance {
	collection := s.IsResourceModelCollection(instance)
	if collection {
		panic(errInvalidResourceInstance)
	}

	v := reflect.ValueOf(instance)
	if v.IsNil() {
		return nil
	}

	m := &resourceModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		f.parseResourceModel(m)
	}
	return i
}

func (s *Schema) parseJoinResourceModel(instance interface{}) *SchemaInstance {
	v := reflect.ValueOf(instance)

	// if instance is not a pointer,
	// wrap it in a pointer
	if v.Type() == s.resourceModelType {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		v = ptr
	}

	if v.Type() != reflect.PtrTo(s.resourceModelType) {
		panic(errInvalidResourceInstance)
	}
	if v.IsNil() {
		return nil
	}

	m := &resourceModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		f.parseJoinResourceModel(m)
	}
	return i
}

func (s *Schema) ParseJsonapiModelCollection(instance interface{}) []*SchemaInstance {
	collection := s.IsJsonapiModelCollection(instance)
	if !collection {
		panic(errInvalidJsonapiCollection)
	}

	v := reflect.ValueOf(instance)
	if v.IsNil() {
		return nil
	}

	var schemaInstances []*SchemaInstance
	for i := 0; i < v.Len(); i++ {
		child := v.Index(i)
		schemaInstance := s.ParseJsonapiModel(child.Interface())
		if schemaInstance != nil {
			schemaInstances = append(schemaInstances, schemaInstance)
		}
	}

	return schemaInstances
}

func (s *Schema) ParseJsonapiModel(instance interface{}) *SchemaInstance {
	collection := s.IsJsonapiModelCollection(instance)
	if collection {
		panic(errInvalidJsonapiInstance)
	}

	v := reflect.ValueOf(instance)
	if v.Type() != reflect.PtrTo(s.jsonapiModelType) {
		panic(errInvalidJsonapiInstance)
	}
	if v.IsNil() {
		return nil
	}

	m := &jsonapiModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		f.parseJsonapiModel(m)
	}
	return i
}

func (s *Schema) parseJoinJsonapiModel(instance interface{}) *SchemaInstance {
	v := reflect.ValueOf(instance)
	if v.Type() != reflect.PtrTo(s.joinJsonapiModelType) {
		panic(errInvalidJoinJsonapiInstance)
	}
	if v.IsNil() {
		return nil
	}

	m := &joinJsonapiModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		f.parseJoinJsonapiModel(m)
	}
	return i
}

func (s *Schema) ParsePGModelCollection(instance interface{}) []*SchemaInstance {
	collection := s.IsPGModelCollection(instance)
	if !collection {
		panic(errInvalidPGCollection)
	}

	v := reflect.ValueOf(instance)
	if v.IsNil() {
		return nil
	}

	var schemaInstances []*SchemaInstance
	for i := 0; i < v.Len(); i++ {
		child := v.Index(i)
		schemaInstance := s.ParsePGModel(child.Interface())
		if schemaInstance != nil {
			schemaInstances = append(schemaInstances, schemaInstance)
		}
	}

	return schemaInstances
}

func (s *Schema) ParsePGModel(instance interface{}) *SchemaInstance {
	collection := s.IsPGModelCollection(instance)
	if collection {
		panic(errInvalidJsonapiInstance)
	}

	v := reflect.ValueOf(instance)
	if v.Type() != reflect.PtrTo(s.pgModelType) {
		panic(errInvalidPGInstance)
	}
	if v.IsNil() {
		return nil
	}

	m := &pgModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		f.parsePGModel(m)
	}
	return i
}

func (s *Schema) parseJoinPGModel(instance interface{}) *SchemaInstance {
	v := reflect.ValueOf(instance)
	if v.Type() != reflect.PtrTo(s.joinPGModelType) {
		panic(errInvalidJoinPGInstance)
	}
	if v.IsNil() {
		return nil
	}

	m := &joinPGModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		f.parseJoinPGModel(m)
	}
	return i
}

func (s *Schema) createInstance() *SchemaInstance {
	i := &SchemaInstance{
		schema: s,
	}
	for _, f := range s.fields {
		i.fields = append(i.fields, f.createInstance())
	}
	return i
}

// UnmarshalJsonapiPayload unmarshals a JSON API payload from a Reader into a
// resource model instance.
// If mustBeWritable is true, only writable fields are set,
// otherwise all fields are set to the values in the payload.
func (s *Schema) UnmarshalJsonapiPayload(in io.Reader, resourceModelInstance interface{}, validate *validator.Validate,
	mustBeWritable bool) (interface{}, error) {
	si := s.ParseResourceModel(resourceModelInstance)

	// parse payload into new jsonapi instance
	jsonapiTargetInstance := s.NewJsonapiModelInstance()
	err := jsonapi.UnmarshalPayload(in, jsonapiTargetInstance)
	if err != nil {
		return nil, err
	}

	val := reflect.ValueOf(jsonapiTargetInstance)
	jmi := &jsonapiModelInstance{
		schema: s,
		value:  &val,
	}

	// copy original Resource model fields to a new target Resource model,
	// applying writable fields from parsed jsonapi model
	target := s.newResourceModelInstance()
	for _, fieldInstance := range si.fields {
		if !mustBeWritable || fieldInstance.parentField().Writable() {
			fieldInstance.parseJsonapiModel(jmi)

			// NOTE: this validates any writable field,
			// regardless if it has actually been set by the user
			if validate != nil {
				err = fieldInstance.validate(validate)
				if err != nil {
					return nil, err
				}
			}
		}
		fieldInstance.applyToResourceModel(target)
	}

	return target.value.Interface(), nil
}

func (s *Schema) newResourceModelInstance() *resourceModelInstance {
	v := reflect.New(s.resourceModelType)
	return &resourceModelInstance{
		schema: s,
		value:  &v,
	}
}

func (s *Schema) newJsonapiModelInstance() *jsonapiModelInstance {
	v := reflect.New(s.jsonapiModelType)
	return &jsonapiModelInstance{
		schema: s,
		value:  &v,
	}
}

func (s *Schema) newJoinJsonapiModelInstance() *joinJsonapiModelInstance {
	v := reflect.New(s.joinJsonapiModelType)
	return &joinJsonapiModelInstance{
		schema: s,
		value:  &v,
	}
}

func (s *Schema) newJoinPGModelInstance() *joinPGModelInstance {
	v := reflect.New(s.joinPGModelType)
	return &joinPGModelInstance{
		schema: s,
		value:  &v,
	}
}

func (s *Schema) newPGModelInstance() *pgModelInstance {
	v := reflect.New(s.pgModelType)
	return &pgModelInstance{
		schema: s,
		value:  &v,
	}
}
