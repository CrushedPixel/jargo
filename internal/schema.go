package internal

import (
	"reflect"
	"errors"
	"crushedpixel.net/jargo/api"
	"gopkg.in/go-playground/validator.v9"
)

// resource model -> resource schema
// resource schema instance <-> jsonapi model, pg model, resource model instances

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

// implements api.Schema
type schema struct {
	name  string // jsonapi member name
	table string // sql table name
	alias string // sql table alias

	fields []field

	resourceModelType reflect.Type
	jsonapiModelType  reflect.Type
	pgModelType       reflect.Type

	// model types to be referenced in relations
	// from other jsonapi/pg models,
	// avoiding infinite recursion
	joinJsonapiModelType reflect.Type
	joinPGModelType      reflect.Type

	// validator to be used
	validator validator.Validate
}

func (s *schema) Name() string {
	return s.name
}

func (s *schema) IsResourceModelCollection(data interface{}) bool {
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

func (s *schema) IsJsonapiModelCollection(data interface{}) bool {
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

func (s *schema) IsPGModelCollection(data interface{}) bool {
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

func (s *schema) NewResourceModelInstance() interface{} {
	return reflect.New(s.resourceModelType).Interface()
}

func (s *schema) NewResourceModelCollection(entries ... interface{}) interface{} {
	l := len(entries)
	val := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(s.resourceModelType)), l, l)
	for i := 0; i < l; i++ {
		val.Index(i).Set(reflect.ValueOf(entries[i]))
	}
	return val.Interface()
}

func (s *schema) NewJsonapiModelInstance() interface{} {
	return reflect.New(s.jsonapiModelType).Interface()
}

func (s *schema) NewPGModelInstance() interface{} {
	return reflect.New(s.pgModelType).Interface()
}

func (s *schema) NewPGModelCollection(entries ... interface{}) interface{} {
	l := len(entries)
	val := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(s.pgModelType)), l, l)
	for i := 0; i < l; i++ {
		val.Index(i).Set(reflect.ValueOf(entries[i]))
	}
	return val.Interface()
}

func (s *schema) ParseResourceModelCollection(instance interface{}) []api.SchemaInstance {
	collection := s.IsResourceModelCollection(instance)
	if !collection {
		panic(errInvalidResourceCollection)
	}

	v := reflect.ValueOf(instance)
	if v.IsNil() {
		return nil
	}

	var schemaInstances []api.SchemaInstance
	for i := 0; i < v.Len(); i++ {
		child := v.Index(i)
		schemaInstance := s.ParseResourceModel(child.Interface())
		if schemaInstance != nil {
			schemaInstances = append(schemaInstances, schemaInstance)
		}
	}

	return schemaInstances
}

func (s *schema) ParseResourceModel(instance interface{}) api.SchemaInstance {
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

func (s *schema) ParseJoinResourceModel(instance interface{}) api.SchemaInstance {
	v := reflect.ValueOf(instance)
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

func (s *schema) ParseJsonapiModelCollection(instance interface{}) []api.SchemaInstance {
	collection := s.IsJsonapiModelCollection(instance)
	if !collection {
		panic(errInvalidJsonapiCollection)
	}

	v := reflect.ValueOf(instance)
	if v.IsNil() {
		return nil
	}

	var schemaInstances []api.SchemaInstance
	for i := 0; i < v.Len(); i++ {
		child := v.Index(i)
		schemaInstance := s.ParseJsonapiModel(child.Interface())
		if schemaInstance != nil {
			schemaInstances = append(schemaInstances, schemaInstance)
		}
	}

	return schemaInstances
}

func (s *schema) ParseJsonapiModel(instance interface{}) api.SchemaInstance {
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

func (s *schema) ParseJoinJsonapiModel(instance interface{}) api.SchemaInstance {
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

func (s *schema) ParsePGModelCollection(instance interface{}) []api.SchemaInstance {
	collection := s.IsPGModelCollection(instance)
	if !collection {
		panic(errInvalidPGCollection)
	}

	v := reflect.ValueOf(instance)
	if v.IsNil() {
		return nil
	}

	var schemaInstances []api.SchemaInstance
	for i := 0; i < v.Len(); i++ {
		child := v.Index(i)
		schemaInstance := s.ParsePGModel(child.Interface())
		if schemaInstance != nil {
			schemaInstances = append(schemaInstances, schemaInstance)
		}
	}

	return schemaInstances
}

func (s *schema) ParsePGModel(instance interface{}) api.SchemaInstance {
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

func (s *schema) ParseJoinPGModel(instance interface{}) api.SchemaInstance {
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

func (s *schema) createInstance() *schemaInstance {
	i := &schemaInstance{
		schema: s,
	}
	for _, f := range s.fields {
		i.fields = append(i.fields, f.createInstance())
	}
	return i
}

// implements api.SchemaInstance
type schemaInstance struct {
	schema *schema
	fields []fieldInstance
}

func (i *schemaInstance) ToResourceModel() interface{} {
	instance := i.schema.newResourceModelInstance()
	for _, f := range i.fields {
		f.applyToResourceModel(instance)
	}
	return instance.value.Interface()
}

func (i *schemaInstance) ToJoinResourceModel() interface{} {
	instance := i.schema.newResourceModelInstance()
	for _, f := range i.fields {
		f.applyToJoinResourceModel(instance)
	}
	return instance.value.Interface()
}

func (i *schemaInstance) ToJsonapiModel() interface{} {
	instance := i.schema.newJsonapiModelInstance()
	for _, f := range i.fields {
		f.applyToJsonapiModel(instance)
	}
	return instance.value.Interface()
}

func (i *schemaInstance) ToJoinJsonapiModel() interface{} {
	instance := i.schema.newJoinJsonapiModelInstance()
	for _, f := range i.fields {
		f.applyToJoinJsonapiModel(instance)
	}
	return instance.value.Interface()
}

func (i *schemaInstance) ToPGModel() interface{} {
	instance := i.schema.newPGModelInstance()
	for _, f := range i.fields {
		f.applyToPGModel(instance)
	}
	return instance.value.Interface()
}

func (i *schemaInstance) ToJoinPGModel() interface{} {
	instance := i.schema.newJoinPGModelInstance()
	for _, f := range i.fields {
		f.applyToJoinPGModel(instance)
	}
	return instance.value.Interface()
}

func (i *schemaInstance) Validate(validate *validator.Validate) error {
	for _, f := range i.fields {
		err := f.validate(validate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *schema) newResourceModelInstance() *resourceModelInstance {
	v := reflect.New(s.resourceModelType)
	return &resourceModelInstance{
		schema: s,
		value:  &v,
	}
}

func (s *schema) newJsonapiModelInstance() *jsonapiModelInstance {
	v := reflect.New(s.jsonapiModelType)
	return &jsonapiModelInstance{
		schema: s,
		value:  &v,
	}
}

func (s *schema) newJoinJsonapiModelInstance() *joinJsonapiModelInstance {
	v := reflect.New(s.joinJsonapiModelType)
	return &joinJsonapiModelInstance{
		schema: s,
		value:  &v,
	}
}

func (s *schema) newJoinPGModelInstance() *joinPGModelInstance {
	v := reflect.New(s.joinPGModelType)
	return &joinPGModelInstance{
		schema: s,
		value:  &v,
	}
}

func (s *schema) newPGModelInstance() *pgModelInstance {
	v := reflect.New(s.pgModelType)
	return &pgModelInstance{
		schema: s,
		value:  &v,
	}
}
