package capsulecd_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"capsulecd/lib"
	"path"
	"os"
)


func TestPopulateSystemConfigFileWithSampleConfigurationFile(t *testing.T) {

	config := capsulecd.NewConfiguration("github")
	config.PopulateSystemConfigFile(path.Join("testdata", "sample_configuration.yml"))

	assert.Equal(t, "sample_test_token", config.SourceGithubAccessToken, "should populate github_access_token")
	assert.Equal(t, "sample_auth_token", config.NpmAuthToken,  "should populate engine_npm_auth_token")
	assert.Equal(t, "sample_pypi_password", config.PypiPassword,  "should populate pypi_password")
	assert.Equal(t, "-----BEGIN RSA PRIVATE KEY-----\nsample_supermarket_key\n-----END RSA PRIVATE KEY-----\n", config.ChefSupermarketKey,  "should correctly base64 decode chef supermarket key")

	assert.Equal(t, "patch", config.EngineVersionBumpType,  "should have correct defaults")

	assert.Empty(t, config.RubygemsApiKey, "should have null value for  rubygems_api_key")
}

func TestPopulateSystemConfigFileWithInvalidConfigurationFile(t *testing.T) {

	config := capsulecd.NewConfiguration("github")
	config.PopulateSystemConfigFile(path.Join("testdata", "incorrect_configuration.yml"))

	assert.Equal(t, "sample_test_token", config.SourceGithubAccessToken, "should populate github_access_token")
	assert.Equal(t, "sample_auth_token", config.NpmAuthToken,  "should populate engine_npm_auth_token")
	assert.Equal(t, "patch", config.EngineVersionBumpType,  "should have correct defaults")
	assert.Empty(t, config.ChefSupermarketKey,  "should have an empty chef supermarket key")
	assert.Empty(t, config.RubygemsApiKey, "should have emtpy value for  rubygems_api_key")
}


func TestLoadEnv(t *testing.T) {



	os.Setenv("CAPSULE_PYPI_PASSWORD", "envPyPiPassword")

	config := capsulecd.NewConfiguration("github")

	assert.Equal(t, "envPyPiPasswordd", config.PypiPassword, "should populate PyPiPassword from environmental variable")

}

//func TestOverridingSampleConfigurationFile(t *testing.T) {
//
//	config := capsulecd.NewConfiguration("github")
//	config.PopulateSystemConfigFile(path.Join("testdata", "incorrect_configuration.yml"))
//
//	assert.Equal(t, "sample_test_token", config.SourceGithubAccessToken, "should populate github_access_token")
//	assert.Equal(t, "sample_auth_token", config.NpmAuthToken,  "should populate engine_npm_auth_token")
//	assert.Equal(t, "patch", config.EngineVersionBumpType,  "should have correct defaults")
//	assert.Empty(t, config.ChefSupermarketKey,  "should have an empty chef supermarket key")
//	assert.Empty(t, config.RubygemsApiKey, "should have emtpy value for  rubygems_api_key")
//}