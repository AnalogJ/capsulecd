package engine_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"capsulecd/lib/config"
	"capsulecd/lib/engine"
	"capsulecd/lib/scm"
	"path"
	"path/filepath"
)

func TestEngineChef(t *testing.T) {
	config.Init()
	config.Set("scm","github")
	config.Set("package_type","chef")

	//githubScm, err := scm.Create()
	//assert.NoError(t, err)
	chefEngine, err := engine.Create()
	assert.NoError(t, err)

	assert.Implements(t, (*engine.Engine)(nil), chefEngine, "should implement the Engine interface")

}

func TestEngineChef_BuildStep(t *testing.T) {
	//dirPath, err := ioutil.TempDir("testdata","")
	//assert.NoError(t, err)

	config.Init()
	config.Set("scm","github")
	config.Set("package_type","chef")
	//config.Set("scm_git_parent_path",dirPath)

	githubScm, err := scm.Create()
	assert.NoError(t, err)
	githubScm.Init(nil)

	chefEngine, err := engine.Create()
	assert.NoError(t, err)

	absPath, aerr := filepath.Abs(path.Join("testdata","chef","cookbook_analogj_test"))
	assert.NoError(t, aerr)

	githubScm.Options().GitLocalPath = absPath
	chefEngine.Init(&githubScm)

	berr := chefEngine.BuildStep()
	assert.NoError(t, berr)
}
