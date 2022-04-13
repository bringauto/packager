package bringauto_git

import "bringauto/modules/bringauto_prerequisites"

func (git *Git) FillDefault(*bringauto_prerequisites.Args) error {
	return nil
}

func (git *Git) FillDynamics(*bringauto_prerequisites.Args) error {
	return nil
}

// CheckPrerequisites
// Function should check if the git can be run and if not it returns error
// (not nil value)
func (git *Git) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	// Git server as cmdline constructor for remote server is not good
	return nil
}
