// +build chef

package engine_test

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/engine"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestEngineChef_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)

	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "chef")
	testConfig.Set("scm_github_access_token","placeholder")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)


	//test
	chefEngine, err := engine.Create("chef", pipelineData, testConfig, githubScm)

	//assert
	require.NoError(t, err)
	require.NotNil(t, chefEngine)
	require.Equal(t, "Other", testConfig.GetString("chef_supermarket_type"), "should load engine defaults")
}

func TestEngineChef_AssembleStep(t *testing.T) {
	//setup
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(parentPath)
	dirPath := path.Join(parentPath, "cookbook_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), dirPath)
	require.NoError(t, cerr)

	pipelineData := new(pipeline.Data)

	pipelineData.GitParentPath = parentPath
	pipelineData.GitLocalPath = dirPath
	testConfig, _ := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "chef")
	testConfig.Set("scm_github_access_token","placeholder")

	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)

	chefEngine, err := engine.Create("chef", pipelineData, testConfig, githubScm)
	require.NoError(t, err)

	//test
	berr := chefEngine.AssembleStep()
	require.NoError(t, berr)

	//assert
	require.True(t, utils.FileExists(path.Join(dirPath, "RakeFile")))
	require.True(t, utils.FileExists(path.Join(dirPath, "Berksfile")))
	require.True(t, utils.FileExists(path.Join(dirPath, ".gitignore")))
	require.True(t, utils.FileExists(path.Join(dirPath, "Gemfile")))
}

func TestEngineChef_AssembleStep_WithMinimalCookbook(t *testing.T) {
	//setup
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(parentPath)
	dirPath := path.Join(parentPath, "cookbook_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "chef", "minimal_cookbook_analogj_test"), dirPath)
	require.NoError(t, cerr)

	pipelineData := new(pipeline.Data)
	pipelineData.GitParentPath = parentPath
	pipelineData.GitLocalPath = dirPath
	testConfig, _ := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "chef")
	testConfig.Set("scm_github_access_token","placeholder")

	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)

	chefEngine, err := engine.Create("chef", pipelineData, testConfig, githubScm)
	require.NoError(t, err)

	//test
	berr := chefEngine.AssembleStep()
	require.NoError(t, berr)

	//assert
	require.True(t, utils.FileExists(path.Join(dirPath, "RakeFile")), "should generate recommended files" )
	require.True(t, utils.FileExists(path.Join(dirPath, "Berksfile")), "should generate recommended files" )
	require.True(t, utils.FileExists(path.Join(dirPath, ".gitignore")), "should generate recommended files" )
	require.True(t, utils.FileExists(path.Join(dirPath, "Gemfile")), "should generate recommended files" )
}

func TestEngineChef_AssembleStep_WithoutMetadata(t *testing.T) {
	//setup
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(parentPath)
	dirPath := path.Join(parentPath, "cookbook_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), dirPath)
	require.NoError(t, cerr)

	os.Remove(path.Join(dirPath, "metadata.rb"))

	pipelineData := new(pipeline.Data)
	absPath, aerr := filepath.Abs(dirPath)
	require.NoError(t, aerr)

	pipelineData.GitLocalPath = absPath
	testConfig, _ := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "chef")
	testConfig.Set("scm_github_access_token","placeholder")

	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)

	chefEngine, err := engine.Create("chef", pipelineData, testConfig, githubScm)
	require.NoError(t, err)

	//test
	berr := chefEngine.AssembleStep()

	//assert
	require.Error(t, berr, "shoule return an error")

}

func TestEngineChef_DependenciesStep(t *testing.T) {
	//setup
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(parentPath)
	dirPath := path.Join(parentPath, "cookbook_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), dirPath)
	require.NoError(t, cerr)

	pipelineData := new(pipeline.Data)
	absPath, aerr := filepath.Abs(dirPath)
	require.NoError(t, aerr)

	pipelineData.GitLocalPath = absPath
	testConfig, _ := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "chef")
	testConfig.Set("scm_github_access_token","placeholder")

	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)

	chefEngine, err := engine.Create("chef", pipelineData, testConfig, githubScm)
	require.NoError(t, err)
	berr := chefEngine.AssembleStep()
	require.NoError(t, berr)

	//test
	terr := chefEngine.DependenciesStep()

	//assert
	require.NoError(t, terr)

}

func TestEngineChef_TestStep(t *testing.T) {
	//setup
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(parentPath)
	dirPath := path.Join(parentPath, "cookbook_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), dirPath)
	require.NoError(t, cerr)

	pipelineData := new(pipeline.Data)
	absPath, aerr := filepath.Abs(dirPath)
	require.NoError(t, aerr)

	pipelineData.GitLocalPath = absPath
	testConfig, _ := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "chef")
	testConfig.Set("scm_github_access_token","placeholder")

	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)

	chefEngine, err := engine.Create("chef", pipelineData, testConfig, githubScm)
	require.NoError(t, err)

	berr := chefEngine.AssembleStep()
	require.NoError(t, berr)

	derr := chefEngine.DependenciesStep()
	require.NoError(t, derr)

	//test
	terr := chefEngine.TestStep()

	//assert
	require.NoError(t, terr)
}


func TestEngineChef_PackageStep(t *testing.T) {
	//setup
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(parentPath)
	dirPath := path.Join(parentPath, "cookbook_analogj_test")

	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), dirPath)
	require.NoError(t, cerr)

	pipelineData := new(pipeline.Data)
	absPath, aerr := filepath.Abs(dirPath)
	require.NoError(t, aerr)

	pipelineData.GitLocalPath = absPath
	testConfig, _ := config.Create()
	testConfig.Set("scm", "github")
	testConfig.Set("package_type", "chef")
	testConfig.Set("scm_github_access_token","placeholder")

	githubScm, err := scm.Create("github", pipelineData, testConfig, nil)
	require.NoError(t, err)

	chefEngine, err := engine.Create("chef", pipelineData, testConfig, githubScm)
	require.NoError(t, err)

	berr := chefEngine.AssembleStep()
	require.NoError(t, berr)

	derr := chefEngine.DependenciesStep()
	require.NoError(t, derr)

	//test
	perr := chefEngine.PackageStep()

	//assert
	require.NoError(t, perr)
}