package bringauto_prerequisites

import (
	"fmt"
	"reflect"
)

const (
	fillDefaultsFuncNameConst       = "FillDefault"
	fillDynamicsFuncNameConst       = "FillDynamic"
	checkPrerequisitesFuncNameConst = "CheckPrerequisites"
)

// Initialize initializes structure K which implements
// interface PrerequisitesInterface. If initialization succeed
// the nil is returned. One of functions from PrerequisitesInterface returns error (non-nil value)
// the error is returned immediately after function that returned error.
//
// functions are called in order as described by PrerequisitesInterface
//
// If the structure K does not implement PrerequisitesInterface the
// nil is returned.
func Initialize[K any](instance *K, args ...any) error {
	instanceTypeKind := reflect.TypeOf(*instance).Kind()
	if instanceTypeKind != reflect.Struct {
		return fmt.Errorf("type is not a struct")
	}

	prerequisitesType := reflect.TypeOf((*PrerequisitesInterface)(nil)).Elem()

	argsValues := prepareArgs(args...)

	if !reflect.TypeOf(instance).Implements(prerequisitesType) {
		return nil
	}

	var err error
	m := reflect.ValueOf(instance)

	if IsEmpty(instance) {
		err = callFunction(&m, fillDefaultsFuncNameConst, argsValues)
		if err != nil {
			return err
		}
	}
	err = callFunction(&m, fillDynamicsFuncNameConst, argsValues)
	if err != nil {
		return err
	}
	err = callFunction(&m, checkPrerequisitesFuncNameConst, argsValues)
	if err != nil {
		return err
	}

	return nil
}

// CreateAndInitialize same as Initialized be it creates an instance of struct K and then
// call Initialize and return initialized instance of K.
func CreateAndInitialize[K any](args ...any) *K {
	var instance K
	err := Initialize(&instance, args...)
	if err != nil {
		panic(err)
	}
	return &instance
}

func callFunction(value *reflect.Value, functionName string, argsValues *[]reflect.Value) error {
	m := value.MethodByName(functionName)
	v := m.Call(*argsValues)
	if len(v) < 1 {
		panic("invalid function parameters")
	}
	errAsInterface := v[0].Interface()
	if errAsInterface != nil {
		return errAsInterface.(error)
	}
	return nil
}

func prepareArgs(args ...any) *[]reflect.Value {
	if len(args) == 1 {
		typeOf := reflect.TypeOf(args[0])
		argsType := reflect.TypeOf((*Args)(nil)).Elem()
		if typeOf.Kind() == reflect.Pointer && argsType.Name() == typeOf.Elem().Name() {
			return &[]reflect.Value{reflect.ValueOf(args[0])}
		}
	}

	reflectArgsValue := Args{variadicArgs: args}
	reflectValue := reflect.ValueOf(&reflectArgsValue)
	return &[]reflect.Value{reflectValue}
}
