package config_test

import (
	"capsulecd/pkg/config"
	"github.com/stretchr/testify/assert"
	"testing"

	"log"
	"os"
	"path"
)

func TestInitShouldCorrectlyInitializeConfiguration(t *testing.T) {

	config.Init()

	assert.Equal(t, "default", config.GetString("package_type"), "should populate package_type with default")
	assert.Equal(t, "default", config.GetString("scm"), "should populate scm with default")
	assert.Equal(t, "default", config.GetString("runner"), "should populate runner with default")

	assert.Equal(t, "patch", config.GetString("engine_version_bump_type"), "should populate runner with default")
	assert.Empty(t, config.GetString("rubygems_api_key"), "should have empty value for rubygems_api_key")
}

func TestEnvVariablesShouldLoadProperly(t *testing.T) {

	os.Setenv("CAPSULE_PYPI_PASSWORD", "env_pypi_password")
	os.Setenv("CAPSULE_RUBYGEMS_API_KEY", "env_rubygems_password")
	os.Setenv("CAPSULE_ENGINE_VERSION_BUMP_TYPE", "major")
	config.Init()

	log.Println(config.AllSettings())
	assert.Equal(t, "env_pypi_password", config.GetString("pypi_password"), "should populate PyPiPassword from environmental variable")
	assert.Equal(t, "env_rubygems_password", config.GetString("rubygems_api_key"), "should populate RubyGems Api Key from environmental variable")
	assert.Equal(t, "major", config.GetString("engine_version_bump_type"), "should populate Engine Version Bump Type from environmental variable")

	os.Unsetenv("CAPSULE_PYPI_PASSWORD")
	os.Unsetenv("CAPSULE_RUBYGEMS_API_KEY")
	os.Unsetenv("CAPSULE_ENGINE_VERSION_BUMP_TYPE")
}

func TestReadConfigWithSampleConfigurationFile(t *testing.T) {

	config.Init()
	config.ReadConfig(path.Join("testdata", "sample_configuration.yml"))

	assert.Equal(t, "sample_test_token", config.GetString("scm_github_access_token"), "should populate scm_github_access_token")
	assert.Equal(t, "sample_auth_token", config.GetString("npm_auth_token"), "should populate engine_npm_auth_token")
	assert.Equal(t, "sample_pypi_password", config.GetString("pypi_password"), "should populate pypi_password")
	assert.Equal(t, "-----BEGIN RSA PRIVATE KEY-----\nsample_supermarket_key\n-----END RSA PRIVATE KEY-----\n", config.GetBase64Decoded("chef_supermarket_key"), "should correctly base64 decode chef supermarket key")
}

func TestReadConfigWithMultipleConfigurationFiles(t *testing.T) {

	config.Init()
	config.ReadConfig(path.Join("testdata", "sample_configuration.yml"))
	config.ReadConfig(path.Join("testdata", "sample_configuration_overrides.yml"))

	assert.Equal(t, "sample_test_token_override", config.GetString("scm_github_access_token"), "should populate scm_github_access_token from overrides file")
	assert.Equal(t, "sample_auth_token_override", config.GetString("npm_auth_token"), "should populate engine_npm_auth_token from overrides file")
	assert.Equal(t, "sample_pypi_password", config.GetString("pypi_password"), "should populate pypi_password")
	assert.Equal(t, "-----BEGIN RSA PRIVATE KEY-----\nsample_supermarket_key\n-----END RSA PRIVATE KEY-----\n", config.GetBase64Decoded("chef_supermarket_key"), "should correctly base64 decode chef supermarket key")
}
