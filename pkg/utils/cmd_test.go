package utils_test

import (
	"capsulecd/pkg/utils"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestBashCmdExec(t *testing.T) {
	t.Parallel()

	//test
	cerr := utils.BashCmdExec("echo 'hello world'", "", "")

	//assert
	assert.NoError(t, cerr)
}

func TestCmdExec_Date(t *testing.T) {
	t.Parallel()

	//test
	cerr := utils.CmdExec("date", []string{}, "", "")

	//assert
	assert.NoError(t, cerr)
}

func TestCmdExec_Echo(t *testing.T) {
	t.Parallel()

	//test
	cerr := utils.CmdExec("echo", []string{"hello", "world"}, "", "")

	//assert
	assert.NoError(t, cerr)
}

func TestCmdExec_Error(t *testing.T) {
	t.Parallel()

	//test
	cerr := utils.CmdExec("/bin/bash", []string{"exit", "1"}, "", "")

	//assert
	assert.Error(t, cerr)
}

func TestCmdExec_WorkingDirRelative(t *testing.T) {
	t.Parallel()

	//test
	cerr := utils.CmdExec("ls", []string{}, "testdata", "")

	//assert
	assert.Error(t, cerr)
}

func TestCmdExec_WorkingDirAbsolute(t *testing.T) {
	t.Parallel()

	//test
	absPath, aerr := filepath.Abs(".")
	cerr := utils.CmdExec("ls", []string{}, absPath, "")

	//assert
	assert.NoError(t, aerr)
	assert.NoError(t, cerr)
}
