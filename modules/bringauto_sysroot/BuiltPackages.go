package bringauto_sysroot

import (
	"encoding/json"
	"os"
	"path"
)

const (
	jsonFileName = "built_packages.json"
)

// Contains built packages in sysroot and have functions for Json encoding and decoding of built
// packages.
type BuiltPackages struct {
	Packages []string
}

func (builtPackages *BuiltPackages) AddToBuiltPackages(packageName string) error {
	builtPackages.Packages = append(builtPackages.Packages, packageName)
	bytes, err := json.Marshal(builtPackages.Packages)
	if err != nil {
		return err
	}
	err = os.WriteFile(path.Join(sysrootDirectoryName, jsonFileName), bytes, 0644)
	return err
}

func (builtPackages *BuiltPackages) UpdateBuiltPackages() error {
	bytes, err := os.ReadFile(path.Join(sysrootDirectoryName, jsonFileName))
	if err != nil {
		return nil
	}

	err = json.Unmarshal(bytes, &builtPackages.Packages)
	return err
}
