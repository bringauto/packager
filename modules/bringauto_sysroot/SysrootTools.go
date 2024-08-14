package bringauto_sysroot

import (
	"bringauto/modules/bringauto_log"
	"os"
	"io"
)

// IsSysrootDirectoryEmpty
// Returns true if specified dir do not exists or exists but is empty, otherwise returns false
func IsSysrootDirectoryEmpty() bool {
	f, err := os.Open(sysrootDirectoryName)
	if err != nil { // The directory do not exists
		return true
	}
	defer f.Close()

	_, err = f.Readdirnames(1)

	if err == io.EOF { // The directory exists, but is empty
		return true
	} else if err != nil {
		bringauto_log.GetLogger().Warn("Cannot read in sysroot directory: %s", err)
	}

	return false
}
