package errors_test

import(
	"testing"
	"capsulecd/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCheckErr_WithoutError(t *testing.T) {
	t.Parallel()

	//assert
	assert.NotPanics(t, func(){
		errors.CheckErr(nil)
	})
}

func TestCheckErr_Error(t *testing.T) {
	t.Parallel()

	//assert
	assert.Panics(t, func(){
		errors.CheckErr(errors.Custom("This is an error"))
	})
}

func TestCustom(t *testing.T){
	t.Parallel()

	//assert
	assert.Implements(t, (*error)(nil), errors.Custom("my error"), "should implement the error interface")
}

func TestErrors(t *testing.T){
	t.Parallel()

	//assert
	assert.Implements(t, (*error)(nil), errors.EngineBuildPackageFailed{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineBuildPackageInvalid{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineDistCredentialsMissing{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineDistPackageError{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineTestDependenciesError{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineTestRunnerError{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineTransformUnavailableStep{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineUnspecifiedError{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.EngineValidateToolError{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmAuthenticationFailed{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmFilesystemError{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmPayloadFormatError{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmPayloadUnsupported{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmUnauthorizedUser{}, "should implement the error interface")
	assert.Implements(t, (*error)(nil), errors.ScmUnspecifiedError{}, "should implement the error interface")
}