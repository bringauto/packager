package bringauto_process

type CmdLineHandlerInterface interface {
	// GenerateCmdLine generates a CMD line arguments.
	// In case of error it must return ([]string{},  <error_instance>).
	// In case of success the CMD line arguments are returned and error is set to nil
	GenerateCmdLine() ([]string, error)
}
