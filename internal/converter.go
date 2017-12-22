package internal

import (
	"reflect"
	"github.com/pkg/errors"
)

const (
	pgModelField      = "JargoPGModel"
	jsonapiModelField = "JargoJsonapiModel"
)

var errNoPGModel = errors.New("instance is not a jargo database model")
var errNoResourceModel = errors.New("instance is not the correct jargo resource model")

// takes a slice of struct pointers or a struct value
func isPGModel(instance reflect.Value) (bool, bool) {
	typ := instance.Type()

	collection := false
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
		collection = true
	}

	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Struct {
		_, ok := typ.Elem().FieldByName(pgModelField)
		return ok, collection
	}

	return false, false
}

// takes a slice of struct pointers or a struct value
func isResourceModel(resource *Resource, instance reflect.Value) (bool, bool) {
	typ := instance.Type()

	collection := false
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
		collection = true
	}

	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Struct {
		return resource == nil || typ.Elem() == resource.modelType, collection
	}

	return false, false
}

// takes a slice of struct pointers or a struct value
func isJsonapiModel(instance reflect.Value) (bool, bool) {
	typ := instance.Type()

	collection := false
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
		collection = true
	}

	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Struct {
		_, ok := typ.Elem().FieldByName(jsonapiModelField)
		return ok, collection
	}

	return false, false
}

// takes a slice of struct pointers or struct pointer as instance value
func (r *Registry) pgModelToResourceModel(resource *Resource, instance reflect.Value) (interface{}, error) {
	ok, collection := isPGModel(instance)
	if !ok {
		return nil, errNoPGModel
	}

	if collection {
		// data is slice of struct pointers
		data := reflect.New(reflect.SliceOf(reflect.PtrTo(resource.modelType))).Elem()
		for i := 0; i < instance.Len(); i++ {
			pgInstance := instance.Index(i)
			modelInstance := reflect.New(resource.modelType)

			err := r.copyPGFieldsToResource(resource, pgInstance, modelInstance)
			if err != nil {
				return nil, err
			}

			data = reflect.Append(data, modelInstance)
		}
		return data.Interface(), nil
	} else {
		modelInstance := reflect.New(resource.modelType)

		err := r.copyPGFieldsToResource(resource, instance, modelInstance)
		if err != nil {
			return nil, err
		}
		return modelInstance.Interface(), nil
	}
}

// takes struct pointer values
func (r *Registry) copyPGFieldsToResource(resource *Resource, source reflect.Value, target reflect.Value) error {
	if source.IsNil() {
		return nil
	}

	for j := 0; j < source.Elem().NumField(); j++ {
		pgField := source.Elem().Field(j)
		modelField := target.Elem().FieldByName(source.Type().Elem().Field(j).Name)
		if !modelField.IsValid() {
			continue
		}

		if ok, _ := isPGModel(pgField); ok {
			// get resource type from model relation field
			res, err := r.getResource(getStructType(modelField.Type()))
			if err != nil {
				return err
			}

			// convert relation field to pg field as well
			resourceModel, err := r.pgModelToResourceModel(res, pgField)
			if err != nil {
				return err
			}
			modelField.Set(reflect.ValueOf(resourceModel))
		} else {
			// attributes are copied over
			modelField.Set(pgField)
		}
	}

	return nil
}

// takes a slice of struct pointers or struct pointer as instance value
func (r *Registry) resourceModelToJsonapiModel(resource *Resource, instance reflect.Value, jsonapiModelType reflect.Type) (interface{}, error) {
	ok, collection := isResourceModel(resource, instance)
	if !ok {
		return nil, errNoResourceModel
	}

	if collection {
		// data is slice of struct pointers
		data := reflect.New(reflect.SliceOf(reflect.PtrTo(jsonapiModelType))).Elem()
		for i := 0; i < instance.Len(); i++ {
			modelInstance := instance.Index(i)
			jsonapiInstance := reflect.New(jsonapiModelType)

			err := r.copyResourceFieldsToJsonapi(modelInstance, jsonapiInstance)
			if err != nil {
				return nil, err
			}
			data = reflect.Append(data, jsonapiInstance)
		}
		return data.Interface(), nil
	} else {
		jsonapiInstance := reflect.New(jsonapiModelType)

		err := r.copyResourceFieldsToJsonapi(instance, jsonapiInstance)
		if err != nil {
			return nil, err
		}
		return jsonapiInstance.Interface(), nil
	}
}

// takes struct pointer values
func (r *Registry) copyResourceFieldsToJsonapi(source reflect.Value, target reflect.Value) error {
	if source.IsNil() {
		return nil
	}

	for j := 0; j < source.Elem().NumField(); j++ {
		modelField := source.Elem().Field(j)
		jsonapiField := target.Elem().FieldByName(source.Type().Elem().Field(j).Name)
		if !jsonapiField.IsValid() {
			continue
		}

		if ok, _ := isResourceModel(nil, jsonapiField); ok {
			// get resource type from model relation field
			res, err := r.getResource(getStructType(modelField.Type()))
			if err != nil {
				return err
			}
			// convert relation field to jsonapi field as well
			resourceModel, err := r.resourceModelToJsonapiModel(res, modelField, res.joinJsonapiModel)
			if err != nil {
				return err
			}
			jsonapiField.Set(reflect.ValueOf(resourceModel))
		} else {
			// attributes are copied over
			jsonapiField.Set(modelField)
		}
	}

	return nil
}

// takes a slice of struct pointers or struct pointer as instance value
func (r *Registry) jsonapiModelToResourceModel(resource *Resource, instance reflect.Value) (interface{}, error) {
	ok, collection := isJsonapiModel(instance)
	if !ok {
		return nil, errNoPGModel
	}

	if collection {
		// data is slice of struct pointers
		data := reflect.New(reflect.SliceOf(reflect.PtrTo(resource.modelType))).Elem()
		for i := 0; i < instance.Len(); i++ {
			jsonapiInstance := instance.Index(i)
			modelInstance := reflect.New(resource.modelType)

			err := r.copyJsonapiFieldsToResource(resource, jsonapiInstance, modelInstance)
			if err != nil {
				return nil, err
			}

			data = reflect.Append(data, modelInstance)
		}
		return data.Interface(), nil
	} else {
		modelInstance := reflect.New(resource.modelType)

		err := r.copyJsonapiFieldsToResource(resource, instance, modelInstance)
		if err != nil {
			return nil, err
		}
		return modelInstance.Interface(), nil
	}
}

// takes struct pointer values
func (r *Registry) copyJsonapiFieldsToResource(resource *Resource, source reflect.Value, target reflect.Value) error {
	if source.IsNil() {
		return nil
	}

	for j := 0; j < source.Elem().NumField(); j++ {
		jsonapiField := source.Elem().Field(j)
		modelField := target.Elem().FieldByName(source.Type().Elem().Field(j).Name)
		if !modelField.IsValid() {
			continue
		}

		if ok, _ := isJsonapiModel(jsonapiField); ok {
			// get resource type from model relation field
			res, err := r.getResource(getStructType(modelField.Type()))
			if err != nil {
				return err
			}

			// convert relation field to pg field as well
			resourceModel, err := r.jsonapiModelToResourceModel(res, jsonapiField)
			if err != nil {
				return err
			}
			modelField.Set(reflect.ValueOf(resourceModel))
		} else {
			// attributes are copied over
			modelField.Set(jsonapiField)
		}
	}

	return nil
}

// takes a slice of struct pointers or struct pointer as instance value
func (r *Registry) resourceModelToPGModel(resource *Resource, instance reflect.Value, pgModelType reflect.Type) (interface{}, error) {
	ok, collection := isResourceModel(resource, instance)
	if !ok {
		return nil, errNoResourceModel
	}

	if collection {
		// data is slice of struct pointers
		data := reflect.New(reflect.SliceOf(reflect.PtrTo(pgModelType))).Elem()
		for i := 0; i < instance.Len(); i++ {
			modelInstance := instance.Index(i)
			pgInstance := reflect.New(pgModelType)

			err := r.copyResourceFieldsToPG(modelInstance, pgInstance)
			if err != nil {
				return nil, err
			}
			data = reflect.Append(data, pgInstance)
		}
		return data.Interface(), nil
	} else {
		pgInstance := reflect.New(pgModelType)

		err := r.copyResourceFieldsToPG(instance, pgInstance)
		if err != nil {
			return nil, err
		}
		return pgInstance.Interface(), nil
	}
}

// takes struct pointer values
func (r *Registry) copyResourceFieldsToPG(source reflect.Value, target reflect.Value) error {
	if source.IsNil() {
		return nil
	}

	for j := 0; j < source.Elem().NumField(); j++ {
		modelField := source.Elem().Field(j)
		pgField := target.Elem().FieldByName(source.Type().Elem().Field(j).Name)
		if !pgField.IsValid() {
			continue
		}

		if ok, _ := isResourceModel(nil, pgField); ok {
			// get resource type from model relation field
			res, err := r.getResource(getStructType(modelField.Type()))
			if err != nil {
				return err
			}
			// convert relation field to pg field as well
			resourceModel, err := r.resourceModelToPGModel(res, modelField, res.joinPGModel)
			if err != nil {
				return err
			}
			pgField.Set(reflect.ValueOf(resourceModel))
		} else {
			// attributes are copied over
			pgField.Set(modelField)
		}
	}

	return nil
}
