package capsulecd

import (
	"log"
	"gopkg.in/yaml.v2"
	"path"
	"os"
	"io/ioutil"
	"encoding/base64"

	"strings"
	"fmt"
	"capsulecd/lib/utils"
	"reflect"
	"strconv"
)


// Override order is as follows:
// 1. defaults
// 2. system config file
// 3. repo config file
// 4. runner overrides
// 5. environment overrides
// 6. CLI overrides
type configuration struct {

	// Cli config, shouldnt be set via environmental variables (will be overridden)
	PackageType 			string `yaml:"package_type"`
	Source 				string `yaml:"source"`
	Runner 				string `yaml:"runner"`
	DryRun 				bool `yaml:"dry_run"`

	// General config
	ConfigPath 			string

	// Source config (any credentials added here should also be added to the spec_helper.rb VCR config)
	SourceGitParentPath 		string `yaml:"source_git_parent_path"`
	SourceGithubApiEndpoint 	string `yaml:"source_github_api_endpoint"`
	SourceGithubWebEndpoint 	string `yaml:"source_github_web_endpoint"`
	SourceGithubAccessToken 	string `yaml:"source_github_access_token"`

	// Runner config
	RunnerPullRequest 		string `yaml:"runner_pull_request"`
	RunnerSha 			string `yaml:"runner_sha"`
	RunnerBranch 			string `yaml:"runner_branch"`
	RunnerCloneUrl 			string `yaml:"runner_clone_url"`
	RunnerRepoFullName 		string `yaml:"runner_repo_full_name"`
	RunnerRepoName 			string `yaml:"runner_repo_name"`

	// Package auth/config (any credentials added here should also be added to the spec_helper.rb VCR config)
	ChefSupermarketUsername 	string `yaml:"chef_supermarket_username"`
	NpmAuthToken 			string `yaml:"npm_auth_token"`
	PypiUsername 			string `yaml:"pypi_username"`
	PypiPassword 			string `yaml:"pypi_password"`
	ChefSupermarketType 		string `yaml:"chef_supermarket_type"`
	ChefSupermarketKey 		string `yaml:"chef_supermarket_key"`
	RubygemsApiKey 			string `yaml:"rubygems_api_key"`

	// Engine config
	EngineDisableTest 		bool `yaml:"engine_disable_test"`
	EngineDisableMinification 	bool `yaml:"engine_disable_minification"`
	EngineDisableLint 		bool `yaml:"engine_disable_lint"`
	EngineDisableCoverage 		bool `yaml:"engine_disable_coverage"`
	EngineDisableCleanup 		bool `yaml:"engine_disable_cleanup"`
	EngineCmdTest 			string `yaml:"engine_cmd_test"`
	EngineCmdMinification 		string `yaml:"engine_cmd_minification"`
	EngineCmdLint 			string `yaml:"engine_cmd_lint"`
	EngineCmdCoverage 		string `yaml:"engine_cmd_coverage"`
	EngineVersionBumpType 		string `yaml:"engine_version_bump_type"`
}

func NewConfiguration(source string) *configuration {
	return &configuration{
		Source: source,
		EngineVersionBumpType: "patch",
		ChefSupermarketType: "Other",
	}
}


func (config *configuration) PopulateRepoConfigFile(repoLocalPath string){

	repoConfigFilePath := path.Join(repoLocalPath, "capsule.yml")

      	config.loadConfigFile(repoConfigFilePath)
	config.loadEnv()
	//populate_runner_overrides
	//populate_env_overrides
	//populate_cli_overrides
	//standardize_settings

}


// The raw parsed configuration file, system level, a repo level configuration file will override settings in this file.
func (config *configuration) PopulateSystemConfigFile(systemConfigFile string){
	config.loadConfigFile(systemConfigFile)
}


// private functions
func (config *configuration) loadConfigFile(configFilePath string) {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		log.Print("The configuration file could not be found. Using defaults")
		return
	}

	log.Print("Loading configuration file: %s", configFilePath)

	config_data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Printf("Error reading configuration file: %s", err)
		return
	}

	config.unmarshalYaml(string(config_data))
}

func (config *configuration) unmarshalYaml(payload string){
	err := yaml.Unmarshal([]byte(payload), &config)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	config.cleanup()
	return
}

func (config *configuration) cleanup() {
	//process the ChefSupermarketKey
	if len(config.ChefSupermarketKey) > 0 {
		key, err := base64.StdEncoding.DecodeString(config.ChefSupermarketKey)
		if err != nil {
			log.Print("Could not decode chef_supermarket_key")
		}
		config.ChefSupermarketKey = string(key)
	}

}

func (config *configuration) loadEnv(){
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")

		if strings.HasPrefix(pair[0], "CAPSULE_") && pair[1] != "" {
			snakecaseSettingName := strings.ToLower(strings.TrimPrefix(pair[0], "CAPSULE_"))
			camelcaseSettingName := utils.SnakeCaseToCamelCase(snakecaseSettingName)
			f := reflect.ValueOf(&config).Elem().FieldByName(camelcaseSettingName)

			if f.IsValid() && f.CanSet() {
				switch fieldKind := f.Kind(); fieldKind {
				case reflect.Bool:
					b, err := strconv.ParseBool(pair[1])
					if err != nil{
						continue
					}
					f.SetBool(b)
				case reflect.String:
					f.SetString(pair[1])
				case reflect.Int:
					i, err := strconv.ParseInt(pair[1], 10, 64)
					if err != nil{
						continue
					}
					f.SetInt(i)
				default:
					fmt.Printf("Unknown field type %s. Skipping.", fieldKind.String())
				}
			} else {
				log.Printf("This setting is not valid: %s. Skipping.", camelcaseSettingName)
			}
		}

	}
}




//func (config *Configuration) populateRepoConfigFile(repoLocalPath string)
//	repoConfigFilePath = path.Join(repoLocalPath, '/capsule.yml')
//	load_config_file(repo_config_file_path)
//	populate_runner_overrides
//	populate_env_overrides
//	populate_cli_overrides
//	standardize_settings
//end


//func (sourceGithub SourceGithub) Configure() {
//	ctx := context.Background()
//	ts := oauth2.StaticTokenSource(
//		&oauth2.Token{AccessToken: "... your access token ..."},
//	)
//	tc := oauth2.NewClient(ctx, ts)
//
//	sourceGithub.client = github.NewClient(tc)
//	return
//}
//
//def chef_supermarket_key
//@chef_supermarket_key.to_s.empty? ? nil : Base64.strict_decode64(@chef_supermarket_key)
//end