package mgr

import (
	"github.com/analogj/capsulecd/pkg/pipeline"
	"net/http"
	"os/exec"
	"github.com/analogj/capsulecd/pkg/errors"
	"path"
	"io/ioutil"
	"os"
	"fmt"
	"github.com/analogj/capsulecd/pkg/config"
	"github.com/analogj/capsulecd/pkg/utils"
	"github.com/analogj/capsulecd/pkg/metadata"
)

func DetectRubyBundler(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	//theres no way to automatically determine if a project was created via Yarn (vs NPM)
	return false
}


type mgrRubyBundler struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}


func (m *mgrRubyBundler) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrRubyBundler) MgrValidateTools() error {
	if _, berr := exec.LookPath("gem"); berr != nil {
		return errors.EngineValidateToolError("gem binary is missing")
	}

	if _, berr := exec.LookPath("bundle"); berr != nil {
		return errors.EngineValidateToolError("bundle binary is missing")
	}
	return nil
}

func (m *mgrRubyBundler) MgrAssembleStep() error {
	// check for/create any required missing folders/files
	if !utils.FileExists(path.Join(m.PipelineData.GitLocalPath, "Gemfile")) {
		ioutil.WriteFile(path.Join(m.PipelineData.GitLocalPath, "Gemfile"),
			[]byte(utils.StripIndent(`source 'https://rubygems.org'
			gemspec`)),
			0644,
		)
	}


	return nil
}

func (m *mgrRubyBundler) MgrDependenciesStep(currentMetadata interface{}, nextMetadata interface{}) error {
	// lets install the gem, and any dependencies
	// http://guides.rubygems.org/make-your-own-gem/

	gemCmd := fmt.Sprintf("gem install %s --ignore-dependencies",
	path.Join(m.PipelineData.GitLocalPath, fmt.Sprintf("%s-%s.gem", nextMetadata.(*metadata.RubyMetadata).Name, nextMetadata.(*metadata.RubyMetadata).Version)))
	if terr := utils.BashCmdExec(gemCmd, m.PipelineData.GitLocalPath, nil, ""); terr != nil {
		return errors.EngineTestDependenciesError("gem install failed. Check gemspec and gem dependencies")
	}

	// install dependencies
	if terr := utils.BashCmdExec("bundle install", m.PipelineData.GitLocalPath, nil, ""); terr != nil {
		return errors.EngineTestDependenciesError("bundle install failed. Check Gemfile")
	}
	return nil
}

func (m *mgrRubyBundler) MgrPackageStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "Gemfile.lock"))
	}
	return nil
}


func (m *mgrRubyBundler) MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.IsSet("rubygems_api_key") {
		return errors.MgrDistCredentialsMissing("Cannot deploy package to rubygems, credentials missing")
	}

	credFile, _ := ioutil.TempFile("", "gem_credentials")
	defer os.Remove(credFile.Name())

	// write the .gem/credentials config jfile.

	credContent := fmt.Sprintf(utils.StripIndent(
		`---
		:rubygems_api_key: %s
		`),
		m.Config.GetString("rubygems_api_key"),
	)

	if _, perr := credFile.Write([]byte(credContent)); perr != nil {
		return perr
	}

	pushCmd := fmt.Sprintf("gem push %s --config-file %s",
		fmt.Sprintf("%s-%s.gem", nextMetadata.(*metadata.RubyMetadata).Name, nextMetadata.(*metadata.RubyMetadata).Version),
		credFile.Name(),
	)
	if derr := utils.BashCmdExec(pushCmd, m.PipelineData.GitLocalPath, nil, ""); derr != nil {
		return errors.MgrDistPackageError("Pushing gem to RubyGems.org using `gem push` failed. Check log for exact error")
	}

	return nil
}
