// +build chef

package engine_test

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/engine"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestEngineChef_BuildStep(t *testing.T) {
	//setup
	parentPath, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(parentPath)
	dirPath := path.Join(parentPath, "cookbook_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), dirPath)
	assert.NoError(t, cerr)

	pipelineData := new(pipeline.Data)
	absPath, aerr := filepath.Abs(dirPath)
	assert.NoError(t, aerr)

	pipelineData.GitLocalPath = absPath
	testConfig, _ := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "chef")

	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	assert.NoError(t, err)

	chefEngine, err := engine.Create("chef", pipelineData, testConfig, githubScm)
	assert.NoError(t, err)

	//test
	berr := chefEngine.BuildStep()
	assert.NoError(t, berr)

	//assert
	assert.True(t, utils.FileExists(path.Join(dirPath, "RakeFile")))
	assert.True(t, utils.FileExists(path.Join(dirPath, "Berksfile")))
	//assert.True(t, utils.FileExists(path.Join(dirPath, ".gitignore"))) //TODO:
	assert.True(t, utils.FileExists(path.Join(dirPath, "Gemfile")))
}

func TestEngineChef_TestStep(t *testing.T) {
	//setup
	parentPath, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(parentPath)
	dirPath := path.Join(parentPath, "cookbook_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), dirPath)
	assert.NoError(t, cerr)

	pipelineData := new(pipeline.Data)
	absPath, aerr := filepath.Abs(dirPath)
	assert.NoError(t, aerr)

	pipelineData.GitLocalPath = absPath
	testConfig, _ := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "chef")

	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	assert.NoError(t, err)

	chefEngine, err := engine.Create("chef", pipelineData, testConfig, githubScm)
	assert.NoError(t, err)

	//test
	berr := chefEngine.BuildStep()
	assert.NoError(t, berr)

	terr := chefEngine.TestStep()
	assert.NoError(t, terr)

	//assert

}
