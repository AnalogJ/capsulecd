package errors_test

import (
	"capsulecd/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckErr_WithoutError(t *testing.T) {
	t.Parallel()

	//assert
	assert.NotPanics(t, func() {
		errors.CheckErr(nil)
	})
}

func TestCheckErr_Error(t *testing.T) {
	t.Parallel()

	//assert
	assert.Panics(t, func() {
		errors.CheckErr(errors.Custom("This is an error"))
	})
}

func TestCustom(t *testing.T) {
	t.Parallel()

	//assert
	assert.Implements(t, (*error)(nil), errors.Custom("my error"), "should implement the error interface")
}

func TestErrors(t *testing.T) {
	t.Parallel()

	//assert
	assert.Implements(t, (*error)(nil), errors.EngineBuildPackageFailed("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineBuildPackageInvalid("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineDistCredentialsMissing("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineDistPackageError("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineTestDependenciesError("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineTestRunnerError("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineTransformUnavailableStep("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineUnspecifiedError("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineValidateToolError("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmAuthenticationFailed("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmFilesystemError("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmPayloadFormatError("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmPayloadUnsupported("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmUnauthorizedUser("test"), "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmUnspecifiedError("test"), "should implement the error interface")
}
