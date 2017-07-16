package utils_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"capsulecd/lib/utils"
	"path/filepath"
)

func TestCmdExec_Date(t *testing.T) {

	cerr := utils.CmdExec("date", []string{},"","")
	assert.NoError(t, cerr)
}

func TestCmdExec_Echo(t *testing.T) {

	cerr := utils.CmdExec("echo", []string{"hello","world"},"","")
	assert.NoError(t, cerr)
}

func TestCmdExec_WorkingDirRelative(t *testing.T) {

	cerr := utils.CmdExec("ls", []string{},"testdata","")
	assert.Error(t, cerr)
}

func TestCmdExec_WorkingDirAbsolute(t *testing.T) {

	absPath, aerr := filepath.Abs(".")
	assert.NoError(t, aerr)

	cerr := utils.CmdExec("ls", []string{}, absPath,"")
	assert.NoError(t, cerr)
}



