package bringauto_prerequisites

// PrerequisitesInterface solves problems with prerequisites
// initialization and consistency checking
type PrerequisitesInterface interface {
	// FillDefault
	// fills up defaults values for the structure elements.
	// Called only if the given structure is considered as empty
	// by IsEmpty function.
	//
	// default values for structure elements are values which
	// can be determined by compile time and can be used as "default".
	//
	// Data filled up by FillDefault must be valid and must be
	// used in further structure usage (no placeholders!).
	//
	FillDefault(args *Args) error

	// FillDynamic
	// serve for a dynamic fill of the user structure.
	// Called after FillDefault
	//
	// Dynamic elements of the structure are elements which values
	// cannot be determined in compile time.
	FillDynamic(args *Args) error

	// CheckPrerequisites
	// check if the prerequisites are met.
	//
	// Function is called after FillDefault and FillDynamic
	CheckPrerequisites(args *Args) error
}
