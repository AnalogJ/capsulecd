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

func TestEngineChef(t *testing.T) {
	config.Init()
	config.Set("scm", "github")
	config.Set("package_type", "chef")

	//githubScm, err := scm.Create()
	//assert.NoError(t, err)
	chefEngine, err := engine.Create()
	assert.NoError(t, err)

	assert.Implements(t, (*engine.Engine)(nil), chefEngine, "should implement the Engine interface")

}

func TestEngineChef_BuildStep(t *testing.T) {
	parentPath, err := ioutil.TempDir("testdata", "")
	assert.NoError(t, err)
	defer os.RemoveAll(parentPath)
	dirPath := path.Join(parentPath, "cookbook_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), dirPath)
	assert.NoError(t, cerr)

	config.Init()
	config.Set("scm", "github")
	config.Set("package_type", "chef")

	githubScm, err := scm.Create()
	assert.NoError(t, err)

	pipelineData := new(pipeline.Data)
	githubScm.Init(pipelineData, nil)

	chefEngine, err := engine.Create()
	assert.NoError(t, err)

	absPath, aerr := filepath.Abs(dirPath)
	assert.NoError(t, aerr)

	pipelineData.GitLocalPath = absPath
	chefEngine.Init(pipelineData, githubScm)

	berr := chefEngine.BuildStep()
	assert.NoError(t, berr)

	assert.True(t, utils.FileExists(path.Join(dirPath, "RakeFile")))
	assert.True(t, utils.FileExists(path.Join(dirPath, "Berksfile")))
	//assert.True(t, utils.FileExists(path.Join(dirPath, ".gitignore"))) //TODO:
	assert.True(t, utils.FileExists(path.Join(dirPath, "Gemfile")))
}

func TestEngineChef_TestStep(t *testing.T) {
	parentPath, err := ioutil.TempDir("testdata", "")
	assert.NoError(t, err)
	defer os.RemoveAll(parentPath)
	dirPath := path.Join(parentPath, "cookbook_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), dirPath)
	assert.NoError(t, cerr)

	config.Init()
	config.Set("scm", "github")
	config.Set("package_type", "chef")

	githubScm, err := scm.Create()
	assert.NoError(t, err)

	pipelineData := new(pipeline.Data)
	githubScm.Init(pipelineData, nil)

	chefEngine, err := engine.Create()
	assert.NoError(t, err)

	absPath, aerr := filepath.Abs(dirPath)
	assert.NoError(t, aerr)

	pipelineData.GitLocalPath = absPath
	chefEngine.Init(pipelineData, githubScm)

	berr := chefEngine.BuildStep()
	assert.NoError(t, berr)

	terr := chefEngine.TestStep()
	assert.NoError(t, terr)
}
