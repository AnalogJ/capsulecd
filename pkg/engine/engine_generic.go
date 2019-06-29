package engine

import (
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/metadata"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"fmt"
	"github.com/Masterminds/semver"
	"io/ioutil"
	"path"
	"strings"
)

type engineGeneric struct {
	engineBase

	Scm             scm.Interface //Interface
	CurrentMetadata *metadata.GenericMetadata
	NextMetadata    *metadata.GenericMetadata
}

func (g *engineGeneric) Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(metadata.GenericMetadata)
	g.NextMetadata = new(metadata.GenericMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault("engine_generic_version_template", `version := "%d.%d.%d"`)
	g.Config.SetDefault("engine_version_metadata_path", "VERSION")
	g.Config.SetDefault("engine_cmd_compile", "echo 'skipping compile'")
	g.Config.SetDefault("engine_cmd_lint", "echo 'skipping lint'")
	g.Config.SetDefault("engine_cmd_fmt", "echo 'skipping fmt'")
	g.Config.SetDefault("engine_cmd_test", "echo 'skipping test'")
	g.Config.SetDefault("engine_cmd_security_check", "echo 'skipping security check'")
	return nil
}

func (g *engineGeneric) GetCurrentMetadata() interface{} {
	return g.CurrentMetadata
}
func (g *engineGeneric) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineGeneric) ValidateTools() error {
	return nil
}

func (g *engineGeneric) AssembleStep() error {
	//validate that the chef metadata.rb file exists

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, g.Config.GetString("engine_version_metadata_path"))) {
		return errors.EngineBuildPackageInvalid(fmt.Sprintf("version file (%s) is required for metadata storage via generic engine", g.Config.GetString("engine_version_metadata_path")))
	}

	// bump up the go package version
	if merr := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath); merr != nil {
		return merr
	}

	if perr := g.populateNextMetadata(); perr != nil {
		return perr
	}

	if nerr := g.writeNextMetadata(g.PipelineData.GitLocalPath); nerr != nil {
		return nerr
	}

	return nil
}

func (g *engineGeneric) PackageStep() error {

	if cerr := utils.GitCommit(g.PipelineData.GitLocalPath, fmt.Sprintf("(v%s) %s", g.NextMetadata.Version, g.Config.GetString("engine_version_bump_msg"))); cerr != nil {
		return cerr
	}
	tagCommit, terr := utils.GitTag(g.PipelineData.GitLocalPath, fmt.Sprintf("v%s", g.NextMetadata.Version), g.Config.GetString("engine_version_bump_msg"))
	if terr != nil {
		return terr
	}

	g.PipelineData.ReleaseCommit = tagCommit
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

//Helpers
func (g *engineGeneric) retrieveCurrentMetadata(gitLocalPath string) error {
	//read VERSION file.
	versionContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, g.Config.GetString("engine_version_metadata_path")))
	if rerr != nil {
		return rerr
	}

	major := 0
	minor := 0
	patch := 0
	template := g.Config.GetString("engine_generic_version_template")
	fmt.Sscanf(strings.TrimSpace(string(versionContent)), template, &major, &minor, &patch)

	g.CurrentMetadata.Version = fmt.Sprintf("%d.%d.%d", major, minor, patch)
	return nil
}

func (g *engineGeneric) populateNextMetadata() error {

	nextVersion, err := g.BumpVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *engineGeneric) writeNextMetadata(gitLocalPath string) error {

	v, nerr := semver.NewVersion(g.NextMetadata.Version)
	if nerr != nil {
		return nerr
	}

	template := g.Config.GetString("engine_generic_version_template")
	versionContent := fmt.Sprintf(template, v.Major(), v.Minor(), v.Patch())

	return ioutil.WriteFile(path.Join(gitLocalPath, g.Config.GetString("engine_version_metadata_path")), []byte(versionContent), 0644)
}
