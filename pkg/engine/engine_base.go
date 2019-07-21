package engine

import (
	"github.com/analogj/capsulecd/pkg/config"
	"github.com/analogj/capsulecd/pkg/errors"
	"github.com/analogj/capsulecd/pkg/pipeline"
	"github.com/analogj/capsulecd/pkg/utils"
	stderrors "errors"
	"fmt"
	"github.com/Masterminds/semver"
)

type engineBase struct {
	Config       config.Interface
	PipelineData *pipeline.Data
}

// default Compile Step.
func (g *engineBase) CompileStep() error {
	if terr := g.ExecuteCmdList("engine_cmd_compile",
		g.PipelineData.GitLocalPath,
		nil,
		"",
		"Compile command (%s) failed. Check log for more details.",
	); terr != nil {
		return terr
	}
	return nil
}

// default Test step
// assumes that the lint and code fmt commands are very similar and that engine_cmd_fmt includes engine_cmd_lint.
func (g *engineBase) TestStep() error {

	//skip the lint commands if disabled
	if !g.Config.GetBool("engine_disable_lint") {
		//run test command
		lintKey := "engine_cmd_lint"
		if g.Config.GetBool("engine_enable_code_mutation") {
			lintKey = "engine_cmd_fmt"
		}

		if terr := g.ExecuteCmdList(lintKey,
			g.PipelineData.GitLocalPath,
			nil,
			"",
			"Lint command (%s) failed. Check log for more details.",
		); terr != nil {
			return terr
		}
	}

	//run test command
	if terr := g.ExecuteCmdList("engine_cmd_test",
		g.PipelineData.GitLocalPath,
		nil,
		"",
		"Test command (%s) failed. Check log for more details.",
	); terr != nil {
		return terr
	}

	//skip the security test commands if disabled
	if !g.Config.GetBool("engine_disable_security_check") {
		//run security check command
		if terr := g.ExecuteCmdList("engine_cmd_security_check",
			g.PipelineData.GitLocalPath,
			nil,
			"",
			"Dependency vulnerability check command (%s) failed. Check log for more details.",
		); terr != nil {
			return terr
		}
	}

	return nil
}

//Helper functions

func (e *engineBase) BumpVersion(currentVersion string) (string, error) {
	v, nerr := semver.NewVersion(currentVersion)
	if nerr != nil {
		return "", nerr
	}

	switch bumpType := e.Config.GetString("engine_version_bump_type"); bumpType {
	case "major":
		return fmt.Sprintf("%d.%d.%d", v.Major()+1, 0, 0), nil
	case "minor":
		return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor()+1, 0), nil
	case "patch":
		return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch()+1), nil
	default:
		return "", stderrors.New("Unknown version bump interval")
	}

}

func (e *engineBase) ExecuteCmdList(configKey string, workingDir string, environ []string, logPrefix string, errorTemplate string) error {
	cmd := e.Config.GetString(configKey)

	// we have to support 2 types of cmds.
	// - simple commands (engine_cmd_compile: 'compile command')
	// and list commands (engine_cmd_compile: - 'compile command' \n - 'compile command 2' \n ..)
	// GetString will return "" if this is a list of commands.

	if cmd != "" {
		//code formatter
		cmdPopulated, aerr := utils.PopulateTemplate(cmd, e.PipelineData)
		if aerr != nil {
			return aerr
		}

		if terr := utils.BashCmdExec(cmdPopulated, workingDir, environ, logPrefix); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf(errorTemplate, cmdPopulated))
		}
	} else {
		cmdList := e.Config.GetStringSlice(configKey)
		if cmdList == nil {
			return nil
		}
		for _, cmd := range cmdList {
			cmdPopulated, aerr := utils.PopulateTemplate(cmd, e.PipelineData)
			if aerr != nil {
				return aerr
			}

			if terr := utils.BashCmdExec(cmdPopulated, workingDir, environ, logPrefix); terr != nil {
				return errors.EngineTestRunnerError(fmt.Sprintf(errorTemplate, cmdPopulated))
			}
		}
	}
	return nil
}
