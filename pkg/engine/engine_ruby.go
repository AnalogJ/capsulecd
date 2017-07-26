package engine

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
)

type rubyMetadata struct {
	Name    string
	Version string
}

type rubyGemspec struct {
	Name    string `json:"name"`
	Version struct {
		Version string `json:"name"`
	} `json:"version"`
}

type engineRuby struct {
	engineBase

	PipelineData    *pipeline.Data
	Scm             scm.Interface //Interface
	CurrentMetadata *rubyMetadata
	NextMetadata    *rubyMetadata
	GemspecPath     string
}

func (g *engineRuby) Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(rubyMetadata)
	g.NextMetadata = new(rubyMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault("engine_cmd_lint", "rubocop --fail-level error")
	g.Config.SetDefault("engine_cmd_fmt", "rubocop --fail-level error --auto-correct")
	g.Config.SetDefault("engine_cmd_test", "rake spec")
	g.Config.SetDefault("engine_cmd_security_check", "bundle audit check --update")
	return nil
}

func (g *engineRuby) ValidateTools() error {
	if _, kerr := exec.LookPath("ruby"); kerr != nil {
		return errors.EngineValidateToolError("ruby binary is missing")
	}

	if _, berr := exec.LookPath("gem"); berr != nil {
		return errors.EngineValidateToolError("gem binary is missing")
	}

	if _, berr := exec.LookPath("bundle"); berr != nil {
		return errors.EngineValidateToolError("bundle binary is missing")
	}

	if _, berr := exec.LookPath("rake"); berr != nil {
		return errors.EngineValidateToolError("rake binary is missing")
	}

	if _, berr := exec.LookPath("rubocop"); berr != nil && !g.Config.GetBool("engine_disable_lint") {
		return errors.EngineValidateToolError("rubocop binary is missing")
	}

	return nil
}

func (g *engineRuby) AssembleStep() error {

	// bump up the version here.
	// since there's no standardized way to bump up the version in the *.gemspec file, we're going to assume that the version
	// is specified in a version file in the lib/<gem_name>/ directory, similar to how the bundler gem does it.
	// ie. bundle gem <gem_name> will create a file: my_gem/lib/my_gem/version.rb with the following contents
	// module MyGem
	//   VERSION = "0.1.0"
	// end
	//
	// Jeweler and Hoe both do something similar.
	// http://yehudakatz.com/2010/04/02/using-gemspecs-as-intended/
	// http://timelessrepo.com/making-ruby-gems
	// http://guides.rubygems.org/make-your-own-gem/
	if merr := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath); merr != nil {
		return merr
	}

	if perr := g.populateNextMetadata(); perr != nil {
		return perr
	}

	if nerr := g.writeNextMetadata(g.PipelineData.GitLocalPath); nerr != nil {
		return nerr
	}

	// check for/create any required missing folders/files
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "Gemfile")) {
		ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, "Gemfile"),
			[]byte(utils.StripIndent(`source 'https://rubygems.org'
			gemspec`)),
			0644,
		)
	}

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "Rakefile")) {
		ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, "Rakefile"),
			[]byte("task :default => :spec"),
			0644,
		)
	}

	os.MkdirAll(path.Join(g.PipelineData.GitLocalPath, "spec"), 0644)

	gitignorePath := path.Join(g.PipelineData.GitLocalPath, ".gitignore")
	if !utils.FileExists(gitignorePath) {
		if err := utils.GitGenerateGitIgnore(g.PipelineData.GitLocalPath, "Ruby"); err != nil {
			return err
		}
	}

	// package the gem, make sure it builds correctly

	gemCmd := fmt.Sprintf("gem build %s", g.GemspecPath)
	if terr := utils.BashCmdExec(gemCmd, g.PipelineData.GitLocalPath, nil, ""); terr != nil {
		return errors.EngineBuildPackageFailed("gem build failed. Check gemspec file and dependencies")
	}

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, fmt.Sprintf("%s-%s.gem", g.NextMetadata.Name, g.NextMetadata.Version))) {
		return errors.EngineBuildPackageFailed(fmt.Sprintf("gem build failed. %s-%s.gem not found", g.NextMetadata.Name, g.NextMetadata.Version))
	}
	return nil
}

func (g *engineRuby) DependenciesStep() error {
	// lets install the gem, and any dependencies
	// http://guides.rubygems.org/make-your-own-gem/

	gemCmd := fmt.Sprintf("gem install %s --ignore-dependencies",
		path.Join(g.PipelineData.GitLocalPath, fmt.Sprintf("%s-%s.gem", g.NextMetadata.Name, g.NextMetadata.Version)))
	if terr := utils.BashCmdExec(gemCmd, g.PipelineData.GitLocalPath, nil, ""); terr != nil {
		return errors.EngineTestDependenciesError("gem install failed. Check gemspec and gem dependencies")
	}

	// install dependencies
	if terr := utils.BashCmdExec("bundle install", g.PipelineData.GitLocalPath, nil, ""); terr != nil {
		return errors.EngineBuildPackageFailed("bundle install failed. Check Gemfile")
	}
	return nil
}

func (g *engineRuby) CompileStep() error {
	return nil
}

func (g *engineRuby) TestStep() error {

	//skip the lint commands if disabled
	if !g.Config.GetBool("engine_disable_lint") {
		//run test command
		lintCmd := g.Config.GetString("engine_cmd_lint")
		if g.Config.GetBool("engine_enable_code_mutation") {
			lintCmd = g.Config.GetString("engine_cmd_fmt")
		}
		if terr := utils.BashCmdExec(lintCmd, g.PipelineData.GitLocalPath, nil, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Lint command (%s) failed. Check log for more details.", lintCmd))
		}
	}
	//skip the lint commands if disabled
	if !g.Config.GetBool("engine_disable_test") {
		//run test command
		testCmd := g.Config.GetString("engine_cmd_test")
		if terr := utils.BashCmdExec(testCmd, g.PipelineData.GitLocalPath, nil, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Test command (%s) failed. Check log for more details.", testCmd))
		}
	}

	//skip the security test commands if disabled
	if !g.Config.GetBool("engine_disable_security_check") {
		//run security check command
		vulCmd := g.Config.GetString("engine_cmd_security_check")
		if terr := utils.BashCmdExec(vulCmd, g.PipelineData.GitLocalPath, nil, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Dependency vulnerability check command (%s) failed. Check log for more details.", vulCmd))
		}
	}
	return nil
}

func (g *engineRuby) PackageStep() error {
	if !g.Config.GetBool("engine_package_keep_lock_file") {
		os.Remove(path.Join(g.PipelineData.GitLocalPath, "Gemfile.lock"))
	}

	if cerr := utils.GitCommit(g.PipelineData.GitLocalPath, fmt.Sprintf("(v%s) Automated packaging of release by CapsuleCD", g.NextMetadata.Version)); cerr != nil {
		return cerr
	}
	tagCommit, terr := utils.GitTag(g.PipelineData.GitLocalPath, fmt.Sprintf("v%s", g.NextMetadata.Version))
	if terr != nil {
		return terr
	}

	g.PipelineData.ReleaseCommit = tagCommit
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *engineRuby) DistStep() error {
	if !g.Config.IsSet("rubygems_api_key") {
		return errors.EngineDistCredentialsMissing("Cannot deploy package to rubygems, credentials missing")
	}

	credFile, _ := ioutil.TempFile("", "gem_credentials")
	defer os.Remove(credFile.Name())

	// write the .gem/credentials config jfile.

	credContent := fmt.Sprintf(utils.StripIndent(
		`---
		:rubygems_api_key: %s
		`),
		g.Config.GetString("rubygems_api_key"),
	)

	if _, perr := credFile.Write([]byte(credContent)); perr != nil {
		return perr
	}

	pushCmd := fmt.Sprintf("gem push %s --config-file %s",
		fmt.Sprintf("%s-%s.gem", g.NextMetadata.Name, g.NextMetadata.Version),
		credFile.Name(),
	)
	if derr := utils.BashCmdExec(pushCmd, g.PipelineData.GitLocalPath, nil, ""); derr != nil {
		return errors.EngineDistPackageError("Pushing gem to RubyGems.org using `gem push` failed. Check log for exact error")
	}

	return nil
}

//private Helpers
func (g *engineRuby) retrieveCurrentMetadata(gitLocalPath string) error {
	//read Gemspec file.
	gemspecFiles, gerr := filepath.Glob(path.Join(gitLocalPath, "/*.gemspec"))
	if gerr != nil {
		return errors.EngineBuildPackageInvalid("*.gemspec file is required to process Ruby gem")
	} else if len(gemspecFiles) == 0 {
		return errors.EngineBuildPackageInvalid("*.gemspec file is required to process Ruby gem")
	}

	g.GemspecPath = gemspecFiles[0]

	gemspecJsonFile, _ := ioutil.TempFile("", "gemspec.json")
	defer os.Remove(gemspecJsonFile.Name())

	//generate a JSON-style YAML file containing the Gemspec data. (still not straight valid JSON).
	//
	gemspecJsonCmd := fmt.Sprintf("ruby -e \"require('yaml'); File.write('%s', YAML::to_json(Gem::Specification::load('%s')))\"",
		gemspecJsonFile.Name(),
		g.GemspecPath,
	)
	if cerr := utils.BashCmdExec(gemspecJsonCmd, "", nil, ""); cerr != nil {
		return errors.EngineBuildPackageFailed(fmt.Sprintf("Command (%s) failed. Check log for more details.", gemspecJsonCmd))
	}

	//Load gemspec JSON file and parse it.
	gemspecJsonContent, rerr := ioutil.ReadFile(gemspecJsonFile.Name())
	if rerr != nil {
		return rerr
	}

	gemspecObj := new(rubyGemspec)
	if uerr := yaml.Unmarshal(gemspecJsonContent, gemspecObj); uerr != nil {
		fmt.Println(string(gemspecJsonContent))
		return uerr
	}

	g.CurrentMetadata.Name = gemspecObj.Name
	g.CurrentMetadata.Version = gemspecObj.Version.Version

	//ensure that there is a lib/GEMNAME/version.rb file.
	versionrbPath := path.Join("lib", gemspecObj.Name, "version.rb")
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, versionrbPath)) {
		return errors.EngineBuildPackageInvalid(
			fmt.Sprintf("version.rb file (%s) is required to process Ruby gem", versionrbPath))
	}
	return nil
}

func (g *engineRuby) populateNextMetadata() error {

	nextVersion, err := g.BumpVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	g.NextMetadata.Name = g.CurrentMetadata.Name
	return nil
}

func (g *engineRuby) writeNextMetadata(gitLocalPath string) error {

	versionrbPath := path.Join(g.PipelineData.GitLocalPath, "lib", g.CurrentMetadata.Name, "version.rb")
	versionrbContent, rerr := ioutil.ReadFile(versionrbPath)
	if rerr != nil {
		return rerr
	}
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	updatedContent := re.ReplaceAllLiteralString(string(versionrbContent), g.NextMetadata.Version)
	return ioutil.WriteFile(versionrbPath, []byte(updatedContent), 0644)
}
