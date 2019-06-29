package config_test

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/pipeline"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func TestConfiguration_init_ShouldCorrectlyInitializeConfiguration(t *testing.T) {
	t.Parallel()

	//test
	testConfig, _ := config.Create()

	//assert
	require.Equal(t, "default", testConfig.GetString("package_type"), "should populate package_type with default")
	require.Equal(t, "default", testConfig.GetString("scm"), "should populate scm with default")
	require.Equal(t, "default", testConfig.GetString("runner"), "should populate runner with default")
	require.Equal(t, "Automated packaging of release by CapsuleCD", testConfig.GetString("engine_version_bump_msg"), "should populate version bump message")
	require.Equal(t, "CapsuleCD", testConfig.GetString("scm_notify_source"), "should populate notify source")
	require.Equal(t, "https://github.com/AnalogJ/capsulecd", testConfig.GetString("scm_notify_target_url"), "should populate notify url")

	require.Equal(t, "patch", testConfig.GetString("engine_version_bump_type"), "should populate runner with default")
	require.Empty(t, testConfig.GetString("rubygems_api_key"), "should have empty value for rubygems_api_key")
}

func TestConfiguration_init_EnvVariablesShouldLoadProperly(t *testing.T) {
	//setup
	os.Setenv("CAPSULE_PYPI_PASSWORD", "env_pypi_password")
	os.Setenv("CAPSULE_RUBYGEMS_API_KEY", "env_rubygems_password")
	os.Setenv("CAPSULE_ENGINE_VERSION_BUMP_TYPE", "major")

	//test
	testConfig, _ := config.Create()

	//assert
	require.Equal(t, "env_pypi_password", testConfig.GetString("pypi_password"), "should populate PyPiPassword from environmental variable")
	require.Equal(t, "env_rubygems_password", testConfig.GetString("rubygems_api_key"), "should populate RubyGems Api Key from environmental variable")
	require.Equal(t, "major", testConfig.GetString("engine_version_bump_type"), "should populate Engine Version Bump Type from environmental variable")

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
	require.Error(t, testConfig.ReadConfig(path.Join("does", "not", "exist.yml")), "should raise an error")
}

func TestConfiguration_ReadConfig_WithSampleConfigurationFile(t *testing.T) {
	t.Parallel()

	//setup
	testConfig, _ := config.Create()

	//test
	testConfig.ReadConfig(path.Join("testdata", "sample_configuration.yml"))
	str64, _ := testConfig.GetBase64Decoded("chef_supermarket_key")
	//assert
	require.Equal(t, "sample_auth_token", testConfig.GetString("npm_auth_token"), "should populate engine_npm_auth_token")
	require.Equal(t, "sample_pypi_password", testConfig.GetString("pypi_password"), "should populate pypi_password")
	require.Equal(t, "-----BEGIN RSA PRIVATE KEY-----\nsample_supermarket_key\n-----END RSA PRIVATE KEY-----\n", str64, "should correctly base64 decode chef supermarket key")
}

func TestConfiguration_ReadConfig_WithAssetConfigurationFile(t *testing.T) {
	t.Parallel()

	//setup
	testConfig, _ := config.Create()

	//test
	testConfig.ReadConfig(path.Join("testdata", "asset_configuration.yml"))

	parsedAssets := new([]pipeline.ScmReleaseAsset)
	err := testConfig.UnmarshalKey("scm_release_assets", parsedAssets)

	//assert
	require.NoError(t, err)
	require.Equal(t, 2, len(*parsedAssets), "should parse scm_release_assets")
	require.Equal(t, "test/path/artifactname.gem", (*parsedAssets)[0].LocalPath, "should parse scm_release_assets")
	require.Equal(t, "artifactname2.gem", (*parsedAssets)[1].ArtifactName, "should parse scm_release_assets")
	require.Equal(t, "", (*parsedAssets)[1].ContentType, "should parse scm_release_assets")

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
	require.Equal(t, "sample_auth_token_override", testConfig.GetString("npm_auth_token"), "should populate engine_npm_auth_token from overrides file")
	require.Equal(t, "sample_pypi_password", testConfig.GetString("pypi_password"), "should populate pypi_password")
	require.Equal(t, "-----BEGIN RSA PRIVATE KEY-----\nsample_supermarket_key\n-----END RSA PRIVATE KEY-----\n", str64, "should correctly base64 decode chef supermarket key")
}

func TestConfiguration_GetBool(t *testing.T) {

	//setup
	os.Setenv("CAPSULE_BOOL_TEST_1", "true")
	os.Setenv("CAPSULE_BOOL_TEST_2", "TRUE")

	//test
	testConfig, _ := config.Create()

	//assert
	require.True(t, testConfig.GetBool("bool_test_1"), "lowercase true in env var should be valid bool")
	require.True(t, testConfig.GetBool("bool_test_2"), "upcase true in env var should be valid bool")
	require.False(t, testConfig.GetBool("bool_test_3"), "unset/default for config should be false")

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
	testConfig.Set("test_key_4", "")                                         //multiline string "this\nis\na multiline\nstring"

	//test
	testKey, terr1 := testConfig.GetBase64Decoded("test_key")
	testKey2, terr2 := testConfig.GetBase64Decoded("test_key_2")
	testKey3, terr3 := testConfig.GetBase64Decoded("test_key_3")
	testKey4, terr4 := testConfig.GetBase64Decoded("test_key_4")

	//assert
	require.Empty(t, testKey)
	require.Error(t, terr1, "should return an error when invalid base64")

	require.Equal(t, "encode this string. ", testKey2, "should correctly decode base64 string")
	require.Nil(t, terr2)

	require.Equal(t, "this\r\nis\r\na multiline\r\nstring", testKey3, "should correctly decode base64 into multiline string")
	require.Nil(t, terr3)

	require.Equal(t, "", testKey4, "should correctly return empty string, if empty string is passed in.")
	require.Nil(t, terr4)
}

func TestConfiguration_GetStringSlice_WithNestedKeys(t *testing.T) {
	t.Parallel()

	//setup
	testConfig, _ := config.Create()

	//test
	testConfig.ReadConfig(path.Join("testdata", "pre_post_step_hook_configuration.yml"))

	pre_scm_init_steps := testConfig.GetStringSlice("scm_init_step.pre")
	post_scm_init_steps := testConfig.GetStringSlice("scm_init_step.post")
	invalid_step := testConfig.GetStringSlice("invalid_step.post")

	//assert
	require.Equal(t, 2, len(pre_scm_init_steps), "should have 2 pre hook commands")
	require.Equal(t, "echo 'pre scm_init_step'", pre_scm_init_steps[0], "should correctly load pre command")
	require.Equal(t, `echo "sdfdsf"`, pre_scm_init_steps[1], "should correctly load pre command with double quotes")
	require.Equal(t, 1, len(post_scm_init_steps), "should have 1 post hook commands")
	require.Nil(t, invalid_step, "invalid step hook should be nil")

}

func TestConfiguration_ListTypeCmd_List(t *testing.T) {
	t.Parallel()

	//setup
	testConfig, _ := config.Create()

	//test
	testConfig.SetDefault("engine_cmd_compile", "echo 'default'")
	testConfig.ReadConfig(path.Join("testdata", "compile_cmd_list_configuration.yml"))

	cmd := testConfig.GetString("engine_cmd_compile")
	cmd_list := testConfig.GetStringSlice("engine_cmd_compile")

	//assert
	require.True(t, testConfig.IsSet("engine_cmd_compile"), "should correctly detect key is populated")
	require.Empty(t, "", cmd, "should return empty when treated as a string")
	require.Equal(t, []string{`echo 'test compile command 1'`, `echo 'test compile command 2'`, `echo 'test compile command 3'`}, cmd_list, "should correctly return entries when treated as a list")
}

func TestConfiguration_ListTypeCmd_Simple(t *testing.T) {
	t.Parallel()

	//setup
	testConfig, _ := config.Create()

	//test
	testConfig.SetDefault("engine_cmd_compile", "echo 'default'")
	testConfig.ReadConfig(path.Join("testdata", "compile_cmd_simple_configuration.yml"))

	cmd := testConfig.GetString("engine_cmd_compile")
	//cmd_list := testConfig.GetStringSlice("engine_cmd_compile")

	//assert
	require.True(t, testConfig.IsSet("engine_cmd_compile"), "should correctly detect key is populated")
	require.Equal(t, `echo "test compile"`, cmd, "list should contain correct command")
}

func TestConfiguration_SetDefault_IsSet_True(t *testing.T) {

	//setup
	testConfig, _ := config.Create()
	testConfig.SetDefault("test_config_key", "test_config_value")

	//test
	isSet := testConfig.IsSet("test_config_key")

	//assert
	require.True(t, isSet, "isSet is true when set via file, flag, default or env")
}
