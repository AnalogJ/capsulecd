package engine

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"encoding/json"
	"fmt"
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
	*engineBase

	PipelineData    *pipeline.Data
	Scm             scm.Interface //Interface
	CurrentMetadata *rubyMetadata
	NextMetadata    *rubyMetadata
	GemspecPath     string
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

	return nil
}

func (g *engineRuby) init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(rubyMetadata)
	g.NextMetadata = new(rubyMetadata)
	return nil
}

func (g *engineRuby) BuildStep() error {

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
			[]byte(`source 'https://rubygems.org'
			gemspec`),
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

	//TODO: add gitignore content.
	//if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, ".gitignore")) {
	//	ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, ".gitignore"),
	//		[]byte(""),
	//		0644,
	//	)
	//}

	// package the gem, make sure it builds correctly

	gemCmd := fmt.Sprintf("gem build %s", g.GemspecPath)
	if terr := utils.BashCmdExec(gemCmd, g.PipelineData.GitLocalPath, ""); terr != nil {
		return errors.EngineBuildPackageFailed("gem build failed. Check gemspec file and dependencies")
	}

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, fmt.Sprintf("%s-%s.gem", g.NextMetadata.Name, g.NextMetadata.Version))) {
		return errors.EngineBuildPackageFailed(fmt.Sprintf("gem build failed. %s-%s.gem not found", g.NextMetadata.Name, g.NextMetadata.Version))
	}
	return nil
}

func (g *engineRuby) TestStep() error {
	// lets install the gem, and any dependencies
	// http://guides.rubygems.org/make-your-own-gem/

	gemCmd := fmt.Sprintf("gem install %s --ignore-dependencies",
		path.Join(g.PipelineData.GitLocalPath, fmt.Sprintf("%s-%s.gem", g.NextMetadata.Name, g.NextMetadata.Version)))
	if terr := utils.BashCmdExec(gemCmd, g.PipelineData.GitLocalPath, ""); terr != nil {
		return errors.EngineTestDependenciesError("gem install failed. Check gemspec and gem dependencies")
	}

	// install dependencies
	if terr := utils.BashCmdExec("bundle install", g.PipelineData.GitLocalPath, ""); terr != nil {
		return errors.EngineBuildPackageFailed("bundle install failed. Check Gemfile")
	}

	//run test command
	var testCmd string
	if g.Config.IsSet("engine_cmd_test") {
		testCmd = g.Config.GetString("engine_cmd_test")
	} else {
		testCmd = "rake spec"
	}
	if terr := utils.BashCmdExec(testCmd, g.PipelineData.GitLocalPath, ""); terr != nil {
		return errors.EngineTestRunnerError(fmt.Sprintf("Test command (%s) failed. Check log for more details.", testCmd))
	}
	return nil
}

func (g *engineRuby) PackageStep() error {
	// commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
	// as they will be included in the commmit and any release artifacts)

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
	credContent := fmt.Sprintf(
		`---
		:rubygems_api_key: %s
		`,
		g.Config.GetString("rubygems_api_key"),
	)

	if _, perr := credFile.Write([]byte(credContent)); perr != nil {
		return perr
	}

	pushCmd := fmt.Sprintf("gem push %s --config-file %s",
		fmt.Sprintf("%s-%s.gem", g.NextMetadata.Name, g.NextMetadata.Version),
		credFile.Name(),
	)
	if derr := utils.BashCmdExec(pushCmd, g.PipelineData.GitLocalPath, ""); derr != nil {
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

	//generate a JSON file containing the Gemspec data.
	gemspecJsonCmd := fmt.Sprintf("require('yaml'); File.write('%s', YAML::to_json(Gem::Specification::load('%s')))",
		gemspecJsonFile.Name(),
		g.GemspecPath,
	)
	if cerr := utils.BashCmdExec(gemspecJsonCmd, "", ""); cerr != nil {
		return cerr
	}

	//Load gemspec JSON file and parse it.
	gemspecJsonContent, rerr := ioutil.ReadFile(gemspecJsonFile.Name())
	if rerr != nil {
		return rerr
	}

	gemspecObj := new(rubyGemspec)
	if uerr := json.Unmarshal(gemspecJsonContent, gemspecObj); uerr != nil {
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
