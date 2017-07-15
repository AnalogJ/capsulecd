package errors

import (
	"log"
	"fmt"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatal("ERROR:", err)
	}
}


// Raised when the config file specifies a hook/override for a step when the type is :repo
type EngineTransformUnavailableStep string
func (str EngineTransformUnavailableStep) Error() string {
	return fmt.Sprintf("Engine Transform Unavailable Step %q", string(str))
}

// Raised when the scm is not recognized
type ScmUnspecifiedError string
func (str ScmUnspecifiedError) Error() string {
	return fmt.Sprintf("Scm Unspecified Error %q", string(str))
}

// Raised when capsule cannot create an authenticated client for the source.
type ScmAuthenticationFailed string
func (str ScmAuthenticationFailed) Error() string {
	return fmt.Sprintf("Scm Authentication Failed %q", string(str))
}

// Raised when there is an error parsing the repo payload format.
type ScmPayloadFormatError string
func (str ScmPayloadFormatError) Error() string {
	return fmt.Sprintf("Scm Payload Format Error %q", string(str))
}

// Raised when a source payload is unsupported/action is invalid
type ScmPayloadUnsupported string
func (str ScmPayloadUnsupported) Error() string {
	return fmt.Sprintf("Scm Payload Unsupported %q", string(str))
}

// Raised when the user who started the packaging is unauthorized (non-collaborator)
type ScmUnauthorizedUser string
func (str ScmUnauthorizedUser) Error() string {
	return fmt.Sprintf("Scm Unauthorized User %q", string(str))
}

// Raised when the package is missing certain required files (ie metadata.rb, package.json, setup.py, etc)
type EngineBuildPackageInvalid string
func (str EngineBuildPackageInvalid) Error() string {
	return fmt.Sprintf("Engine Build Package Invalid %q", string(str))
}

// Raised when the source could not be compiled or build for any reason
type EngineBuildPackageFailed string
func (str EngineBuildPackageFailed) Error() string {
	return fmt.Sprintf("Engine Build Package Failed %q", string(str))
}

// Raised when package dependencies fail to install correctly.
type EngineTestDependenciesError string
func (str EngineTestDependenciesError) Error() string {
	return fmt.Sprintf("Engine Test Dependencies Error %q", string(str))
}

// Raised when the package test runner fails
type EngineTestRunnerError string
func (str EngineTestRunnerError) Error() string {
	return fmt.Sprintf("Engine Test Runner Error %q", string(str))
}

// Raised when credentials required to upload/deploy new package are missing.
type EngineReleaseCredentialsMissing string
func (str EngineReleaseCredentialsMissing) Error() string {
	return fmt.Sprintf("Engine Release Credentials Missing %q", string(str))
}

// Raised when an error occurs while uploading package.
type EngineReleasePackageError string
func (str EngineReleasePackageError) Error() string {
	return fmt.Sprintf("Engine Release Package Error %q", string(str))
}