package errors

import (
	"log"
	"fmt"
	"errors"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatal("ERROR:", err)
	}
}

func Custom(args string) error {
	return errors.New(args)
}

// Raised when there is an issue with the filesystem for scm checkout
type ScmFilesystemError string
func (str ScmFilesystemError) Error() string {
	return fmt.Sprintf("ScmFilesystemError: %q", string(str))
}

// Raised when the scm is not recognized
type ScmUnspecifiedError string
func (str ScmUnspecifiedError) Error() string {
	return fmt.Sprintf("ScmUnspecifiedError: %q", string(str))
}

// Raised when capsule cannot create an authenticated client for the source.
type ScmAuthenticationFailed string
func (str ScmAuthenticationFailed) Error() string {
	return fmt.Sprintf("ScmAuthenticationFailed: %q", string(str))
}

// Raised when there is an error parsing the repo payload format.
type ScmPayloadFormatError string
func (str ScmPayloadFormatError) Error() string {
	return fmt.Sprintf("ScmPayloadFormatError: %q", string(str))
}

// Raised when a source payload is unsupported/action is invalid
type ScmPayloadUnsupported string
func (str ScmPayloadUnsupported) Error() string {
	return fmt.Sprintf("ScmPayloadUnsupported: %q", string(str))
}

// Raised when the user who started the packaging is unauthorized (non-collaborator)
type ScmUnauthorizedUser string
func (str ScmUnauthorizedUser) Error() string {
	return fmt.Sprintf("ScmUnauthorizedUser: %q", string(str))
}

// Raised when the config file specifies a hook/override for a step when the type is :repo
type EngineTransformUnavailableStep string
func (str EngineTransformUnavailableStep) Error() string {
	return fmt.Sprintf("EngineTransformUnavailableStep: %q", string(str))
}

// Raised when the environment is missing a required tool/binary
type EngineValidateToolError string
func (str EngineValidateToolError) Error() string {
	return fmt.Sprintf("EngineValidateToolError: %q", string(str))
}

// Raised when the engine is not recognized
type EngineUnspecifiedError string
func (str EngineUnspecifiedError) Error() string {
	return fmt.Sprintf("EngineUnspecifiedError: %q", string(str))
}

// Raised when the package is missing certain required files (ie metadata.rb, package.json, setup.py, etc)
type EngineBuildPackageInvalid string
func (str EngineBuildPackageInvalid) Error() string {
	return fmt.Sprintf("EngineBuildPackageInvalid: %q", string(str))
}

// Raised when the source could not be compiled or build for any reason
type EngineBuildPackageFailed string
func (str EngineBuildPackageFailed) Error() string {
	return fmt.Sprintf("EngineBuildPackageFailed: %q", string(str))
}

// Raised when package dependencies fail to install correctly.
type EngineTestDependenciesError string
func (str EngineTestDependenciesError) Error() string {
	return fmt.Sprintf("EngineTestDependenciesError: %q", string(str))
}

// Raised when the package test runner fails
type EngineTestRunnerError string
func (str EngineTestRunnerError) Error() string {
	return fmt.Sprintf("EngineTestRunnerError: %q", string(str))
}

// Raised when credentials required to upload/deploy new package are missing.
type EngineDistCredentialsMissing string
func (str EngineDistCredentialsMissing) Error() string {
	return fmt.Sprintf("EngineDistCredentialsMissing: %q", string(str))
}

// Raised when an error occurs while uploading package.
type EngineDistPackageError string
func (str EngineDistPackageError) Error() string {
	return fmt.Sprintf("EngineDistPackageError: %q", string(str))
}