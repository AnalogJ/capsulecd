package mgr

import (
	"capsulecd/pkg/pipeline"
	"net/http"
	"capsulecd/pkg/config"
)

// Create mock using:
// mockgen -source=pkg/mgr/interface.go -destination=pkg/mgr/mock/mock_mgr.go
type Interface interface {

	Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error

	// Validate that required executables are available for the following build/test/package/etc steps
	MgrValidateTools() error

	// Assemble the package contents
	// Validate that any required files (like dependency management files) exist
	// Create any recommended optional/missing files we can in the structure.
	// CAN NOT override
	// REQUIRES pipelineData.GitLocalPath
	MgrAssembleStep() error

	// Validate & download dependencies for this package.
	// Generate *.lock files for dependencies (should be deleted in MgrPackageStep if necessary)
	// CAN override
	// REQUIRES pipelineData.GitLocalPath
	// REQUIRES CurrentMetadata
	// REQUIRES NextMetadata
	MgrDependenciesStep(currentMetadata interface{}, nextMetadata interface{}) error

	// Commit any local changes and create a git tag. Nothing should be pushed to remote repository yet.
	// Make sure you remove any unnecessary files from the repo before making the commit
	// CAN NOT override
	// REQUIRES pipelineData.GitLocalPath
	// REQUIRES NextMetadata
	// USES mgr_keep_lock_file
	MgrPackageStep(currentMetadata interface{}, nextMetadata interface{}) error

	// Push the release to the package repository (ie. npm, chef supermarket, rubygems)
	// Should validate any required credentials are specified.
	// CAN override
	// REQUIRES pipelineData.GitLocalPath
	// REQUIRES NextMetadata
	// USES chef_supermarket_username
	// USES chef_supermarket_key
	// USES npm_auth_token
	// USES pypi_repository
	// USES pypi_username
	// USES pypi_password
	// USES rubygems_api_key
	MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error
}