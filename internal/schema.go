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

func (s *schema) IsResourceModelCollection(data interface{}) (bool, error) {
	typ := reflect.ValueOf(data).Type()
	collection := false
	if typ.Kind() == reflect.Slice {
		collection = true
		typ = typ.Elem()
	}

	if typ != reflect.PtrTo(s.resourceModelType) {
		if collection {
			return false, errInvalidResourceCollection
		} else {
			return false, errInvalidResourceInstance
		}
	}

	return collection, nil
}

func (s *schema) IsJsonapiModelCollection(data interface{}) (bool, error) {
	typ := reflect.ValueOf(data).Type()
	collection := false
	if typ.Kind() == reflect.Slice {
		collection = true
		typ = typ.Elem()
	}

	if typ != reflect.PtrTo(s.jsonapiModelType) {
		if collection {
			return false, errInvalidJsonapiCollection
		} else {
			return false, errInvalidJsonapiInstance
		}
	}

	return collection, nil
}

func (s *schema) IsPGModelCollection(data interface{}) (bool, error) {
	typ := reflect.ValueOf(data).Type()
	collection := false
	if typ.Kind() == reflect.Slice {
		collection = true
		typ = typ.Elem()
	}

	if typ != reflect.PtrTo(s.pgModelType) {
		if collection {
			return false, errInvalidPGCollection
		} else {
			return false, errInvalidPGInstance
		}
	}

	return collection, nil
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

func (s *schema) ParseResourceModelCollection(instance interface{}) ([]api.SchemaInstance, error) {
	collection, err := s.IsResourceModelCollection(instance)
	if err != nil {
		return nil, err
	}
	if !collection {
		return nil, errInvalidResourceCollection
	}

	v := reflect.ValueOf(instance)
	if v.IsNil() {
		return nil, nil
	}

	var schemaInstances []api.SchemaInstance
	for i := 0; i < v.Len(); i++ {
		child := v.Index(i)
		schemaInstance, err := s.ParseResourceModel(child.Interface())
		if err != nil {
			return nil, err
		}
		if schemaInstance != nil {
			schemaInstances = append(schemaInstances, schemaInstance)
		}
	}

	return schemaInstances, nil
}

func (s *schema) ParseResourceModel(instance interface{}) (api.SchemaInstance, error) {
	collection, err := s.IsResourceModelCollection(instance)
	if err != nil {
		return nil, err
	}
	if collection {
		return nil, errInvalidResourceInstance
	}

	v := reflect.ValueOf(instance)
	if v.IsNil() {
		return nil, nil
	}

	m := &resourceModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		err := f.parseResourceModel(m)
		if err != nil {
			return nil, err
		}
	}
	return i, nil
}

func (s *schema) ParseJoinResourceModel(instance interface{}) (api.SchemaInstance, error) {
	v := reflect.ValueOf(instance)
	if v.Type() != reflect.PtrTo(s.resourceModelType) {
		return nil, errInvalidResourceInstance
	}
	if v.IsNil() {
		return nil, nil
	}

	m := &resourceModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		err := f.parseJoinResourceModel(m)
		if err != nil {
			return nil, err
		}
	}
	return i, nil
}

func (s *schema) ParseJsonapiModelCollection(instance interface{}) ([]api.SchemaInstance, error) {
	collection, err := s.IsJsonapiModelCollection(instance)
	if err != nil {
		return nil, err
	}
	if !collection {
		return nil, errInvalidJsonapiCollection
	}

	v := reflect.ValueOf(instance)
	if v.IsNil() {
		return nil, nil
	}

	var schemaInstances []api.SchemaInstance
	for i := 0; i < v.Len(); i++ {
		child := v.Index(i)
		schemaInstance, err := s.ParseJsonapiModel(child.Interface())
		if err != nil {
			return nil, err
		}
		if schemaInstance != nil {
			schemaInstances = append(schemaInstances, schemaInstance)
		}
	}

	return schemaInstances, nil
}

func (s *schema) ParseJsonapiModel(instance interface{}) (api.SchemaInstance, error) {
	collection, err := s.IsJsonapiModelCollection(instance)
	if err != nil {
		return nil, err
	}
	if collection {
		return nil, errInvalidJsonapiInstance
	}

	v := reflect.ValueOf(instance)
	if v.Type() != reflect.PtrTo(s.jsonapiModelType) {
		return nil, errInvalidJsonapiInstance
	}
	if v.IsNil() {
		return nil, nil
	}

	m := &jsonapiModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		err := f.parseJsonapiModel(m)
		if err != nil {
			return nil, err
		}
	}
	return i, nil
}

func (s *schema) ParseJoinJsonapiModel(instance interface{}) (api.SchemaInstance, error) {
	v := reflect.ValueOf(instance)
	if v.Type() != reflect.PtrTo(s.joinJsonapiModelType) {
		return nil, errInvalidJoinJsonapiInstance
	}
	if v.IsNil() {
		return nil, nil
	}

	m := &joinJsonapiModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		err := f.parseJoinJsonapiModel(m)
		if err != nil {
			return nil, err
		}
	}
	return i, nil
}

func (s *schema) ParsePGModelCollection(instance interface{}) ([]api.SchemaInstance, error) {
	collection, err := s.IsPGModelCollection(instance)
	if err != nil {
		return nil, err
	}
	if !collection {
		return nil, errInvalidPGCollection
	}

	v := reflect.ValueOf(instance)
	if v.IsNil() {
		return nil, nil
	}

	var schemaInstances []api.SchemaInstance
	for i := 0; i < v.Len(); i++ {
		child := v.Index(i)
		schemaInstance, err := s.ParsePGModel(child.Interface())
		if err != nil {
			return nil, err
		}
		if schemaInstance != nil {
			schemaInstances = append(schemaInstances, schemaInstance)
		}
	}

	return schemaInstances, nil
}

func (s *schema) ParsePGModel(instance interface{}) (api.SchemaInstance, error) {
	collection, err := s.IsPGModelCollection(instance)
	if err != nil {
		return nil, err
	}
	if collection {
		return nil, errInvalidJsonapiInstance
	}

	v := reflect.ValueOf(instance)
	if v.Type() != reflect.PtrTo(s.pgModelType) {
		return nil, errInvalidPGInstance
	}
	if v.IsNil() {
		return nil, nil
	}

	m := &pgModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		err := f.parsePGModel(m)
		if err != nil {
			return nil, err
		}
	}
	return i, nil
}

func (s *schema) ParseJoinPGModel(instance interface{}) (api.SchemaInstance, error) {
	v := reflect.ValueOf(instance)
	if v.Type() != reflect.PtrTo(s.joinPGModelType) {
		return nil, errInvalidJoinPGInstance
	}
	if v.IsNil() {
		return nil, nil
	}

	m := &joinPGModelInstance{
		schema: s,
		value:  &v,
	}

	i := s.createInstance()
	for _, f := range i.fields {
		err := f.parseJoinPGModel(m)
		if err != nil {
			return nil, err
		}
	}
	return i, nil
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

func (i *schemaInstance) ToResourceModel() (interface{}, error) {
	instance := i.schema.newResourceModelInstance()
	for _, f := range i.fields {
		err := f.applyToResourceModel(instance)
		if err != nil {
			return nil, err
		}
	}
	return instance.value.Interface(), nil
}

func (i *schemaInstance) ToJoinResourceModel() (interface{}, error) {
	instance := i.schema.newResourceModelInstance()
	for _, f := range i.fields {
		err := f.applyToJoinResourceModel(instance)
		if err != nil {
			return nil, err
		}
	}
	return instance.value.Interface(), nil
}

func (i *schemaInstance) ToJsonapiModel() (interface{}, error) {
	instance := i.schema.newJsonapiModelInstance()
	for _, f := range i.fields {
		err := f.applyToJsonapiModel(instance)
		if err != nil {
			return nil, err
		}
	}
	return instance.value.Interface(), nil
}

func (i *schemaInstance) ToJoinJsonapiModel() (interface{}, error) {
	instance := i.schema.newJoinJsonapiModelInstance()
	for _, f := range i.fields {
		err := f.applyToJoinJsonapiModel(instance)
		if err != nil {
			return nil, err
		}
	}
	return instance.value.Interface(), nil
}

func (i *schemaInstance) ToPGModel() (interface{}, error) {
	instance := i.schema.newPGModelInstance()
	for _, f := range i.fields {
		err := f.applyToPGModel(instance)
		if err != nil {
			return nil, err
		}
	}
	return instance.value.Interface(), nil
}

func (i *schemaInstance) ToJoinPGModel() (interface{}, error) {
	instance := i.schema.newJoinPGModelInstance()
	for _, f := range i.fields {
		err := f.applyToJoinPGModel(instance)
		if err != nil {
			return nil, err
		}
	}
	return instance.value.Interface(), nil
}

func (i *schemaInstance) Validate() error {
	for _, f := range i.fields {
		err := f.validate()
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
