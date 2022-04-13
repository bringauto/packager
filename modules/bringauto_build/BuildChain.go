package bringauto_build

// BuildChain represents a chain of commands needed
// to by executed exactly after each other
type BuildChain struct {
	Chain []CMDLineInterface
}

// GenerateCommands for a given build chain.
// Commands must be executable by "bash"
func (build *BuildChain) GenerateCommands() []string {
	var commandList []string
	for _, value := range build.Chain {
		commandList = append(commandList, value.ConstructCMDLine()...)
	}
	return commandList
}
