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
	"os"
	"os/exec"
	"path"
	"strings"
	"capsulecd/pkg/metadata"
	"log"
)

type engineGolang struct {
	engineBase

	Scm             scm.Interface //Interface
	CurrentMetadata *metadata.GolangMetadata
	NextMetadata    *metadata.GolangMetadata
	GoPath          string
}

func (g *engineGolang) Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(metadata.GolangMetadata)
	g.NextMetadata = new(metadata.GolangMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault("engine_cmd_compile", "go build $(go list ./cmd/...)")
	g.Config.SetDefault("engine_cmd_lint", "gometalinter.v2 --errors --vendor --deadline=3m ./...")
	g.Config.SetDefault("engine_cmd_fmt", "go fmt $(go list ./... | grep -v /vendor/)")
	g.Config.SetDefault("engine_cmd_test", "go test $(glide novendor)")
	g.Config.SetDefault("engine_cmd_security_check", "exit 0") //TODO: update when there's a dependency checker for Golang/Glide

	var scmDomain string
	if g.Config.GetString("scm") == "bitbucket" {
		scmDomain = "bitbucket.org"
	} else {
		scmDomain = "github.com"
	}

	g.Config.SetDefault("engine_golang_package_path", fmt.Sprintf("%s/%s",scmDomain, strings.ToLower(g.Config.GetString("scm_repo_full_name"))))

	//TODO: figure out why setting the GOPATH workspace is causing the tools to timeout.
	// golang recommends that your in-development packages are in the GOPATH and glide requires it to do glide install.
	// the problem with this is that for somereason gometalinter (and the underlying linting tools) take alot longer
	// to run, and hit the default deadline limit ( --deadline=30s).
	// we can have multiple workspaces in the gopath by separating them with colon (:), but this timeout is nasty if not required.
	//TODO: g.GoPath root will not be deleted (its the parent of GitParentPath), figure out if we can do this automatically.
	g.GoPath = g.PipelineData.GitParentPath
	os.Setenv("GOPATH", fmt.Sprintf("%s:%s", os.Getenv("GOPATH"), g.GoPath))

	packagePathPrefix := path.Dir(g.Config.GetString("engine_golang_package_path")) //strip out the repo name.
	// customize the git parent path for Golang Engine
	g.PipelineData.GitParentPath = path.Join(g.PipelineData.GitParentPath, "src", packagePathPrefix)
	os.MkdirAll(g.PipelineData.GitParentPath, 0666)

	return nil
}

func (g *engineGolang) GetCurrentMetadata() interface{} {
	return g.CurrentMetadata
}
func (g *engineGolang) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineGolang) ValidateTools() error {
	if _, kerr := exec.LookPath("go"); kerr != nil {
		return errors.EngineValidateToolError("go binary is missing")
	}

	if _, kerr := exec.LookPath("gometalinter.v2"); kerr != nil {
		return errors.EngineValidateToolError("gometalinter.v2 binary is missing")
	}

	return nil
}

func (g *engineGolang) AssembleStep() error {
	//validate that the chef metadata.rb file exists

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "pkg", "version", "version.go")) {
		return errors.EngineBuildPackageInvalid("pkg/version/version.go file is required to process Go library")
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

func (g *engineGolang) CompileStep() error {
	//cmd directory is optional. check if it exists first.
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "cmd")) {
		log.Println("Warning: cmd directory is missing, custom compile command required.")
	}

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

// we cant use the default test step because linter and fmt are very differnt cmds.
func (g *engineGolang) TestStep() error {
	// go test -v $(go list ./... | grep -v /vendor/)
	// gofmt -s -l $(bash find . -name "*.go" | grep -v vendor | uniq)

	//TODO: the package must be in the GOPATH for this to work correctly.
	//http://craigwickesser.com/2015/02/golang-cmd-with-custom-environment/
	//http://www.ryanday.net/2012/10/01/installing-go-and-gopath/
	//

	//skip the lint commands if disabled
	if !g.Config.GetBool("engine_disable_lint") {
		//run lint command
		if terr := g.ExecuteCmdList("engine_cmd_lint",
			g.PipelineData.GitLocalPath,
			nil,
			"",
			"Lint command (%s) failed. Check log for more details.",
		); terr != nil {
			return terr
		}

		if g.Config.GetBool("engine_enable_code_mutation") {
			//code formatter
			if terr := g.ExecuteCmdList("engine_cmd_fmt",
				g.PipelineData.GitLocalPath,
				nil,
				"",
				"Format command (%s) failed. Check log for more details.",
			); terr != nil {
				return terr
			}
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

func (g *engineGolang) PackageStep() error {
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

func (g *engineGolang) customGopathEnv() []string {
	currentEnv := os.Environ()
	updatedEnv := []string{fmt.Sprintf("GOPATH=%s", g.GoPath)}

	for i := range currentEnv {
		if !strings.HasPrefix(currentEnv[i], "GOPATH=") { //add all environmental variables that are not GOPATH
			updatedEnv = append(updatedEnv, currentEnv[i])
		}
	}
	return updatedEnv
}

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
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
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
	f, err := parser.ParseFile(fset, "", string(versionContent), parser.ParseComments)
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
		if gen.Tok == token.CONST || gen.Tok == token.VAR {
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
		if gen.Tok == token.CONST || gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				valSpec := spec.(*ast.ValueSpec)
				if strings.ToLower(valSpec.Names[0].Name) == "version" {
					//found the version variable.
					valSpec.Values[0].(*ast.BasicLit).Value = fmt.Sprintf(`"%s"`, version)
					return list, nil
				}
			}
		}
	}
	return nil, errors.EngineBuildPackageFailed("Could not set the version in pkg/version/version.go")
}
