package utils_test

import (
	"capsulecd/lib/utils"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestCmdExec_Date(t *testing.T) {
	t.Parallel()

	cerr := utils.CmdExec("date", []string{}, "", "")
	assert.NoError(t, cerr)
}

func TestCmdExec_Echo(t *testing.T) {
	t.Parallel()

	cerr := utils.CmdExec("echo", []string{"hello", "world"}, "", "")
	assert.NoError(t, cerr)
}

func TestCmdExec_Error(t *testing.T) {
	t.Parallel()

	cerr := utils.CmdExec("/bin/bash", []string{"exit", "1"}, "", "")
	assert.Error(t, cerr)
}

func TestCmdExec_WorkingDirRelative(t *testing.T) {
	t.Parallel()

	cerr := utils.CmdExec("ls", []string{}, "testdata", "")
	assert.Error(t, cerr)
}

func TestCmdExec_WorkingDirAbsolute(t *testing.T) {
	t.Parallel()

	absPath, aerr := filepath.Abs(".")
	assert.NoError(t, aerr)

	cerr := utils.CmdExec("ls", []string{}, absPath, "")
	assert.NoError(t, cerr)
}
