package bringauto_prerequisites

import "reflect"

// IsEmpty
// Check if the structure is empty.
// Empty structure is a structure which has no explicit initializers.
// Following structures instances are considered empty
//	type A struct {
//		a int
//		b int
//	}
//	var structA A
//	structAA := A{}
//
// It returns true if structure is empty and false if structure is not empty
//
func IsEmpty[K any](arg *K) bool {
	var k K
	return reflect.DeepEqual(k, *arg)
}
