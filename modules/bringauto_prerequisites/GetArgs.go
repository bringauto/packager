package bringauto_prerequisites

import (
	"fmt"
	"reflect"
)

// Args represents arguments passed from Initialize or CreateAndInitialize function.
//
// Let's assume that you want to pass arguments to the FillDynamic function when
// called from Initialize.
// In order to maintain code readability we define structure that holds our arguments:
// Arguments struct implementation
//	type Arguments struct {
//		a string
//		b int
//		c string
//	}
// FillDynamic Implementation:
//	type NiceStruct struct {}
//	func (args *NiceStruct) FillDynamic(args *Args) error {
//		var cmdArgs Arguments
//		bringauto_prerequisites.GetArgs(args, &cmdArgs)
//		fmt.Println(cmdArgs.a)
//		fmt.Println(cmdArgs.b)
//		fmt.Println(cmdArgs.c)
//	}
// Structure initialization:
//	niceStruct := bringauto_prerequisites.CreateAndInitialize("S1", 15, "S2")
// It prints
//	S1
//	15
//	S2
//
type Args struct {
	variadicArgs []any
}

// GetArgs returns arguments set to dataOut structure.
// Types of variadic arguments from Args::variadicArgs must be in same order
// as types in the structure type K.
//
// Variadic arguments is linear list of interface{}. Each element from variadic argument list
// represents specific type. Let T[i] is type of the i-th element from variadic argument list.
//
// Let K is a Go struct type.
//
// If VAL len is not zero then K and VAL must have exact number of fields.
// If VAL len is zero then the dataOut is initialized by K{}: *dataOut = K{}
//
// Let T_K[i] is Type of the i-th element of structure K.
// Then for each i T_K[i] == T[i] otherwise the panic will raise.
func GetArgs[K any](filler *Args, dataOut *K) {
	if len(filler.variadicArgs) == 0 {
		var _t K
		*dataOut = _t
		return
	}

	dataType := reflect.TypeOf(dataOut).Elem()
	dataValue := reflect.ValueOf(dataOut).Elem()

	argsFiledCount := dataType.NumField()
	if len(filler.variadicArgs) != argsFiledCount {
		panic("cannot initialize. Invalid number of parameters")
	}

	for i := 0; i < argsFiledCount; i++ {
		argField := dataType.Field(i)
		argsType := reflect.TypeOf(filler.variadicArgs[i])
		if argField.Type.Name() != argsType.Name() {
			panic(fmt.Errorf("invalid type! Expected '%s', got '%s'", argField.Type.Name(), argsType.Name()))
		}
	}

	for i := 0; i < argsFiledCount; i++ {
		ddValue := dataValue.Field(i)
		if !ddValue.CanSet() {
			panic(fmt.Errorf("field %s::%s cannot be set", dataType.Name(), ddValue.Type().Name()))
		}
		ddValue.Set(reflect.ValueOf(filler.variadicArgs[i]))
	}
}

func prepareVariadicArgs(args ...any) *Args {
	return &Args{args}
}
