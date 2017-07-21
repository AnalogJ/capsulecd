package config_test

import (
	"capsulecd/pkg/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestConfiguration_init_ShouldCorrectlyInitializeConfiguration(t *testing.T) {
	t.Parallel()

	//test
	testConfig, _ := config.Create()

	//assert
	assert.Equal(t, "default", testConfig.GetString("package_type"), "should populate package_type with default")
	assert.Equal(t, "default", testConfig.GetString("scm"), "should populate scm with default")
	assert.Equal(t, "default", testConfig.GetString("runner"), "should populate runner with default")

	assert.Equal(t, "patch", testConfig.GetString("engine_version_bump_type"), "should populate runner with default")
	assert.Empty(t, testConfig.GetString("rubygems_api_key"), "should have empty value for rubygems_api_key")
}

func TestConfiguration_init_EnvVariablesShouldLoadProperly(t *testing.T) {
	//setup
	os.Setenv("CAPSULE_PYPI_PASSWORD", "env_pypi_password")
	os.Setenv("CAPSULE_RUBYGEMS_API_KEY", "env_rubygems_password")
	os.Setenv("CAPSULE_ENGINE_VERSION_BUMP_TYPE", "major")

	//test
	testConfig, _ := config.Create()

	//assert
	assert.Equal(t, "env_pypi_password", testConfig.GetString("pypi_password"), "should populate PyPiPassword from environmental variable")
	assert.Equal(t, "env_rubygems_password", testConfig.GetString("rubygems_api_key"), "should populate RubyGems Api Key from environmental variable")
	assert.Equal(t, "major", testConfig.GetString("engine_version_bump_type"), "should populate Engine Version Bump Type from environmental variable")

	//teardown
	os.Unsetenv("CAPSULE_PYPI_PASSWORD")
	os.Unsetenv("CAPSULE_RUBYGEMS_API_KEY")
	os.Unsetenv("CAPSULE_ENGINE_VERSION_BUMP_TYPE")
}

func TestConfiguration_ReadConfig_InvalidFilePath(t *testing.T) {
	t.Parallel()
	//setup
	testConfig, _ := config.Create()

	//assert
	assert.Error(t, testConfig.ReadConfig(path.Join("does", "not", "exist.yml")), "should raise an error")
}

func TestConfiguration_ReadConfig_WithSampleConfigurationFile(t *testing.T) {
	t.Parallel()

	//setup
	testConfig, _ := config.Create()

	//test
	testConfig.ReadConfig(path.Join("testdata", "sample_configuration.yml"))
	str64, _ := testConfig.GetBase64Decoded("chef_supermarket_key")
	//assert
	assert.Equal(t, "sample_test_token", testConfig.GetString("scm_github_access_token"), "should populate scm_github_access_token")
	assert.Equal(t, "sample_auth_token", testConfig.GetString("npm_auth_token"), "should populate engine_npm_auth_token")
	assert.Equal(t, "sample_pypi_password", testConfig.GetString("pypi_password"), "should populate pypi_password")
	assert.Equal(t, "-----BEGIN RSA PRIVATE KEY-----\nsample_supermarket_key\n-----END RSA PRIVATE KEY-----\n", str64, "should correctly base64 decode chef supermarket key")
}

func TestConfiguration_ReadConfig_WithMultipleConfigurationFiles(t *testing.T) {
	t.Parallel()

	//setup
	testConfig, _ := config.Create()

	//test
	testConfig.ReadConfig(path.Join("testdata", "sample_configuration.yml"))
	testConfig.ReadConfig(path.Join("testdata", "sample_configuration_overrides.yml"))
	str64, _ := testConfig.GetBase64Decoded("chef_supermarket_key")

	//assert
	assert.Equal(t, "sample_test_token_override", testConfig.GetString("scm_github_access_token"), "should populate scm_github_access_token from overrides file")
	assert.Equal(t, "sample_auth_token_override", testConfig.GetString("npm_auth_token"), "should populate engine_npm_auth_token from overrides file")
	assert.Equal(t, "sample_pypi_password", testConfig.GetString("pypi_password"), "should populate pypi_password")
	assert.Equal(t, "-----BEGIN RSA PRIVATE KEY-----\nsample_supermarket_key\n-----END RSA PRIVATE KEY-----\n", str64, "should correctly base64 decode chef supermarket key")
}

func TestConfiguration_GetBool(t *testing.T) {

	//setup
	os.Setenv("CAPSULE_BOOL_TEST_1", "true")
	os.Setenv("CAPSULE_BOOL_TEST_2", "TRUE")

	//test
	testConfig, _ := config.Create()

	//assert
	assert.True(t, testConfig.GetBool("bool_test_1"), "lowercase true in env var should be valid bool")
	assert.True(t, testConfig.GetBool("bool_test_2"), "upcase true in env var should be valid bool")
	assert.False(t, testConfig.GetBool("bool_test_3"), "unset/default for config should be false")

	//teardown
	os.Unsetenv("CAPSULE_BOOL_TEST_1")
	os.Unsetenv("CAPSULE_BOOL_TEST_2")
}

func TestConfiguration_GetBase64Decoded_WithInvalidData(t *testing.T) {
	t.Parallel()

	//setup
	testConfig, _ := config.Create()
	testConfig.Set("test_key", "invalidBase64_encoding")
	testConfig.Set("test_key_2", "ZW5jb2RlIHRoaXMgc3RyaW5nLiA=")             //"encode this string. "
	testConfig.Set("test_key_3", "dGhpcw0KaXMNCmEgbXVsdGlsaW5lDQpzdHJpbmc=") //multiline string "this\nis\na multiline\nstring"
	testConfig.Set("test_key_4", "") //multiline string "this\nis\na multiline\nstring"

	//test
	testKey, terr1 := testConfig.GetBase64Decoded("test_key")
	testKey2, terr2 := testConfig.GetBase64Decoded("test_key_2")
	testKey3, terr3 := testConfig.GetBase64Decoded("test_key_3")
	testKey4, terr4 := testConfig.GetBase64Decoded("test_key_4")

	//assert
	assert.Empty(t, testKey)
	assert.Error(t, terr1, "should return an error when invalid base64")

	assert.Equal(t, "encode this string. ", testKey2, "should correctly decode base64 string")
	assert.Nil(t, terr2)

	assert.Equal(t, "this\r\nis\r\na multiline\r\nstring", testKey3, "should correctly decode base64 into multiline string")
	assert.Nil(t, terr3)

	assert.Equal(t, "", testKey4, "should correctly return empty string, if empty string is passed in.")
	assert.Nil(t, terr4)
}
