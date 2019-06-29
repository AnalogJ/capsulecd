package engine

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/metadata"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type enginePython struct {
	engineBase

	Scm             scm.Interface //Interface
	CurrentMetadata *metadata.PythonMetadata
	NextMetadata    *metadata.PythonMetadata
}

func (g *enginePython) Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(metadata.PythonMetadata)
	g.NextMetadata = new(metadata.PythonMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault("pypi_repository", "https://upload.pypi.org/legacy/")
	g.Config.SetDefault("engine_cmd_compile", "echo 'skipping compile'")
	g.Config.SetDefault("engine_cmd_lint", "find . -name '*.py' -exec pylint -E '{}' +")
	g.Config.SetDefault("engine_cmd_fmt", "find . -name '*.py' -exec pylint -E '{}' +") //TODO: replace with pycodestyle/pep8
	g.Config.SetDefault("engine_cmd_test", "tox")
	g.Config.SetDefault("engine_cmd_security_check", "safety check -r requirements.txt")
	g.Config.SetDefault("engine_version_metadata_path", "VERSION")
	return nil
}

func (g *enginePython) GetCurrentMetadata() interface{} {
	return g.CurrentMetadata
}
func (g *enginePython) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *enginePython) ValidateTools() error {
	if _, kerr := exec.LookPath("tox"); kerr != nil {
		return errors.EngineValidateToolError("tox binary is missing")
	}

	if _, kerr := exec.LookPath("pylint"); kerr != nil && !g.Config.GetBool("engine_disable_lint") {
		return errors.EngineValidateToolError("pylint binary is missing")
	}

	if _, berr := exec.LookPath("python"); berr != nil {
		return errors.EngineValidateToolError("python binary is missing")
	}

	if _, berr := exec.LookPath("safety"); berr != nil && !g.Config.GetBool("engine_disable_security_check") {
		return errors.EngineValidateToolError("safety binary is missing")
	}

	return nil
}

func (g *enginePython) AssembleStep() error {
	//validate that the python setup.py file exists
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "setup.py")) {
		return errors.EngineBuildPackageInvalid("setup.py file is required to process Python package")
	}

	// check for/create required VERSION file
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, g.Config.GetString("engine_version_metadata_path"))) {
		ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, g.Config.GetString("engine_version_metadata_path")),
			[]byte("0.0.0"),
			0644,
		)
	}

	// bump up the version here.
	// since there's no standardized way to bump up the version in the setup.py file, we're going to assume that the version
	// is specified in plain text VERSION file in the root of the source repository. This can be configured via engine_version_metadata_path
	// this is option #4 in the python packaging guide:
	// https://packaging.python.org/en/latest/single_source_version/#single-sourcing-the-version
	//
	// additional packaging structures, like those listed below, may also be supported in the future.
	// http://stackoverflow.com/a/7071358/1157633

	if merr := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath); merr != nil {
		return merr
	}

	if perr := g.populateNextMetadata(); perr != nil {
		return perr
	}

	if nerr := g.writeNextMetadata(g.PipelineData.GitLocalPath); nerr != nil {
		return nerr
	}

	// make sure the package testing manager is available.
	// there is a standardized way to test packages (python setup.py tests), however for automation tox is preferred
	// because of virtualenv and its support for multiple interpreters.
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "tox.ini")) {
		toxIniContent := utils.StripIndent(`# Tox (http://tox.testrun.org/) is a tool for running tests
			# in multiple virtualenvs. This configuration file will run the
			# test suite on all supported python versions. To use it, "pip install tox"
			# and then run "tox" from this directory.
			[tox]
			envlist = py27
			usedevelop = True

			# you may want to change this default test command
			# http://tox.readthedocs.io/en/latest/example/basic.html#integration-with-setup-py-test-command
			[testenv]
			commands = python setup.py test
			deps =
			  -rrequirements.txt
			`)

		ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, "tox.ini"),
			[]byte(toxIniContent),
			0644,
		)
	}

	os.MkdirAll(path.Join(g.PipelineData.GitLocalPath, "tests"), 0644)

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "tests", "__init__.py")) {
		ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, "tests", "__init__.py"),
			[]byte(""),
			0644,
		)
	}

	gitignorePath := path.Join(g.PipelineData.GitLocalPath, ".gitignore")
	if !utils.FileExists(gitignorePath) {
		if err := utils.GitGenerateGitIgnore(g.PipelineData.GitLocalPath, "Python"); err != nil {
			return err
		}
	}
	return nil
}

// use default Compile step
//func (g *enginePython) CompileStep() error { }

// used default Test step
//func (g *enginePython) TestStep() error { }

func (g *enginePython) PackageStep() error {
	os.RemoveAll(path.Join(g.PipelineData.GitLocalPath, ".tox")) //remove .tox folder.

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

//private Helpers

func (g *enginePython) retrieveCurrentMetadata(gitLocalPath string) error {
	//read metadata.json file.
	versionContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, g.Config.GetString("engine_version_metadata_path")))
	if rerr != nil {
		return rerr
	}
	g.CurrentMetadata.Version = strings.TrimSpace(string(versionContent))
	return nil
}

func (g *enginePython) populateNextMetadata() error {

	nextVersion, err := g.BumpVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *enginePython) writeNextMetadata(gitLocalPath string) error {
	return ioutil.WriteFile(path.Join(gitLocalPath, g.Config.GetString("engine_version_metadata_path")), []byte(g.NextMetadata.Version), 0644)
}
