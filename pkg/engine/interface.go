package engine

import (
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/config"
	"capsulecd/pkg/scm"
)

type Interface interface {
	init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error

	// Validate that required executables are available for the following build/test/package/etc steps
	ValidateTools() error

	// Assemble the package contents
	// Validate that any required files (like dependency management files) exist
	// Create any recommended optional/missing files we can in the structure. (.gitignore, etc)
	// Read & Bump the version in the metadata file(s)
	// MUST set CurrentMetadata
	// MUST set NextMetadata
	// REQUIRES pipelineData.GitLocalPath
	AssembleStep() error

	// Validate & download dependencies for this package.
	// Generate *.lock files for dependencies (should be deleted in PackageStep if necessary)
	// REQUIRES pipelineData.GitLocalPath
	// REQUIRES CurrentMetadata
	// REQUIRES NextMetadata
	DependenciesStep() error

	// Validate code syntax & execute test runner
	// Run linter
	// Run unit tests
	// Generate coverage reports
	// USES engine_disable_test
	// USES engine_disable_lint
	// USES engine_enable_code_mutation - allows CapsuleCD to modify code using linting tools (only available on some systems)
	// USES engine_cmd_lint
	// USES engine_cmd_test
	TestStep() error

	// Commit any local changes and create a git tag. Nothing should be pushed to remote repository yet.
	// Once the commit is done, push the package to the package repository.
	// Make sure you remove any unnecessary files from the repo before making the commit
	// MUST set ReleaseCommit
	// MUST set ReleaseVersion
	// REQUIRES pipelineData.GitLocalPath
	// REQUIRES NextMetadata
	// USES engine_package_keep_lock_file
	PackageStep() error


	DistStep() error

	DocumentationStep() error
}

