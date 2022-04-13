package bringauto_repository

import "bringauto/modules/bringauto_package"

// RepositoryInterface is an interface for every package repository.
// If you want to implement
type RepositoryInterface interface {
	// CopyToRepository copy package files represented by sourceDir to the repository.
	// Each repository has a different semantics for managing structure of th repository.
	//
	// Repository must not change the package name represented by pack.CreatePackageName()
	CopyToRepository(pack bringauto_package.Package, sourceDir string) error
}
