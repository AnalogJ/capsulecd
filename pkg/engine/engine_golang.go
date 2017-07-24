package engine

import (
	"bytes"
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/scm"
	"capsulecd/pkg/utils"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
	"os"
)

type golangMetadata struct {
	Version string
}
type engineGolang struct {
	engineBase

	PipelineData    *pipeline.Data
	Scm             scm.Interface //Interface
	CurrentMetadata *golangMetadata
	NextMetadata    *golangMetadata
}

func (g *engineGolang) Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(golangMetadata)
	g.NextMetadata = new(golangMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault("engine_cmd_compile", "go build $(go list ./cmd/...)")
	g.Config.SetDefault("engine_cmd_lint", "gometalinter.v1 ./...")
	g.Config.SetDefault("engine_cmd_fmt", "go fmt $(go list ./... | grep -v /vendor/)")
	g.Config.SetDefault("engine_cmd_test", "go test $(glide novendor)")
	g.Config.SetDefault("engine_cmd_security_check", "exit 0") //TODO: update when there's a dependency checker for Golang/Glide

	return nil
}

func (g *engineGolang) ValidateTools() error {
	if _, kerr := exec.LookPath("go"); kerr != nil {
		return errors.EngineValidateToolError("go binary is missing")
	}

	if _, kerr := exec.LookPath("gometalinter.v1"); kerr != nil {
		return errors.EngineValidateToolError("gometalinter.v1 binary is missing")
	}

	return nil
}

func (g *engineGolang) AssembleStep() error {
	//validate that the chef metadata.rb file exists

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "pkg", "version", "version.go")) {
		return errors.EngineBuildPackageInvalid("pkg/version/version.go file is required to process Go library")
	}

	//we only support glide as a Go dependency manager right now. Should be easy to add additional ones though.
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "glide.yaml")) {
		return errors.EngineBuildPackageInvalid("glide.yml file is required to process Go library")
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

	gitignorePath := path.Join(g.PipelineData.GitLocalPath, ".gitignore")
	if !utils.FileExists(gitignorePath) {
		if err := utils.GitGenerateGitIgnore(g.PipelineData.GitLocalPath, "Go"); err != nil {
			return err
		}
	}

	return nil
}

func (g *engineGolang) DependenciesStep() error {
	//TODO: check if glide will complain if the checkout directory isnt the same as the GOPATH
	// the library has already been downloaded. lets make sure all its dependencies are available.
	if cerr := utils.CmdExec("glide", []string{"install"}, g.PipelineData.GitLocalPath, ""); cerr != nil {
		return errors.EngineTestDependenciesError("glide install failed. Check dependencies")
	}

	return nil
}

func (g *engineGolang) CompileStep() error {
	if g.Config.GetBool("engine_disable_compile") {
		//cmd directory is optional. check if it exists first.
		if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "cmd")) {
			return nil
		}

		//code formatter
		compileCmd := g.Config.GetString("engine_cmd_compile")
		if terr := utils.BashCmdExec(compileCmd, g.PipelineData.GitLocalPath, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Compile command (%s) failed. Check log for more details.", compileCmd))
		}
	}
	return nil
}

func (g *engineGolang) TestStep() error {
	// go test -v $(go list ./... | grep -v /vendor/)
	// gofmt -s -l $(bash find . -name "*.go" | grep -v vendor | uniq)

	//skip the lint commands if disabled
	if !g.Config.GetBool("engine_disable_lint") {
		//run test command
		lintCmd := g.Config.GetString("engine_cmd_lint")
		if terr := utils.BashCmdExec(lintCmd, g.PipelineData.GitLocalPath, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Lint command (%s) failed. Check log for more details.", lintCmd))
		}

		if g.Config.GetBool("engine_enable_code_mutation") {
			//code formatter
			fmtCmd := g.Config.GetString("engine_cmd_fmt")
			if terr := utils.BashCmdExec(fmtCmd, g.PipelineData.GitLocalPath, ""); terr != nil {
				return errors.EngineTestRunnerError(fmt.Sprintf("Format command (%s) failed. Check log for more details.", fmtCmd))
			}
		}
	}

	//skip the test commands if disabled
	if !g.Config.GetBool("engine_disable_test") {
		//run test command
		testCmd :=  g.Config.GetString("engine_cmd_test")
		if terr := utils.BashCmdExec(testCmd, g.PipelineData.GitLocalPath, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Test command (%s) failed. Check log for more details.", testCmd))
		}
	}

	//skip the security test commands if disabled
	if !g.Config.GetBool("engine_disable_security_check") {
		//run security check command
		// no Golang security check known for dependencies.
		//code formatter
		vulCmd := g.Config.GetString("engine_cmd_security_check")
		if terr := utils.BashCmdExec(vulCmd, g.PipelineData.GitLocalPath, ""); terr != nil {
			return errors.EngineTestRunnerError(fmt.Sprintf("Format command (%s) failed. Check log for more details.", vulCmd))
		}
	}
	return nil
}

func (g *engineGolang) PackageStep() error {
	if !g.Config.GetBool("engine_package_keep_lock_file") {
		os.Remove(path.Join(g.PipelineData.GitLocalPath, "glide.lock"))
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

func (g *engineGolang) DistStep() error {

	// no real packaging for golang.
	// libraries are stored in version control.
	return nil
}

//private Helpers

func (g *engineGolang) retrieveCurrentMetadata(gitLocalPath string) error {

	versionContent, rerr := ioutil.ReadFile(path.Join(g.PipelineData.GitLocalPath, "pkg", "version", "version.go"))
	if rerr != nil {
		return rerr
	}

	//Oh.My.God.

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", string(versionContent), 0)
	if err != nil {
		return err
	}

	version, verr := g.parseGoVersion(f.Decls)
	if verr != nil {
		return verr
	}

	g.CurrentMetadata.Version = version
	return nil
}

func (g *engineGolang) populateNextMetadata() error {

	nextVersion, err := g.BumpVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	return nil
}

func (g *engineGolang) writeNextMetadata(gitLocalPath string) error {
	versionPath := path.Join(g.PipelineData.GitLocalPath, "pkg", "version", "version.go")
	versionContent, rerr := ioutil.ReadFile(versionPath)
	if rerr != nil {
		return rerr
	}

	//Oh.My.God.

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", string(versionContent), 0)
	if err != nil {
		return err
	}

	decls, serr := g.setGoVersion(f.Decls, g.NextMetadata.Version)
	if serr != nil {
		return serr
	}
	f.Decls = decls

	//write the version file again.
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return err
	}

	return ioutil.WriteFile(versionPath, buf.Bytes(), 0644)
}

func (g *engineGolang) parseGoVersion(list []ast.Decl) (string, error) {
	//find version declaration (uppercase or lowercase)
	for _, decl := range list {
		gen := decl.(*ast.GenDecl)
		if gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				valSpec := spec.(*ast.ValueSpec)
				if strings.ToLower(valSpec.Names[0].Name) == "version" {
					//found the version variable.
					return strings.Trim(valSpec.Values[0].(*ast.BasicLit).Value, "\"'"), nil
				}
			}
		}
	}
	return "", errors.EngineBuildPackageFailed("Could not retrieve the version from pkg/version/version.go")
}

func (g *engineGolang) setGoVersion(list []ast.Decl, version string) ([]ast.Decl, error) {
	//find version declaration (uppercase or lowercase)
	for _, decl := range list {
		gen := decl.(*ast.GenDecl)
		if gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				valSpec := spec.(*ast.ValueSpec)
				if strings.ToLower(valSpec.Names[0].Name) == "version" {
					//found the version variable.
					valSpec.Values[0].(*ast.BasicLit).Value = version
					return list, nil
				}
			}
		}
	}
	return nil, errors.EngineBuildPackageFailed("Could not set the version in pkg/version/version.go")
}
