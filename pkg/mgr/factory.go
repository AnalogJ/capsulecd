package mgr

import (
	"github.com/analogj/capsulecd/pkg/pipeline"
	"net/http"
	"github.com/analogj/capsulecd/pkg/errors"
	"fmt"
	"github.com/analogj/capsulecd/pkg/config"
)

func Create(mgrType string, pipelineData *pipeline.Data, config config.Interface, client *http.Client) (Interface, error) {

	var mgr Interface

	switch mgrType {
	//empty/generic package manager. Noop.
	case "generic":
		mgr = new(mgrGeneric)

	//chef dependency managers
	case "berkshelf":
		mgr = new(mgrChefBerkshelf)

	//golang dependency managers
	case "dep":
		mgr = new(mgrGolangDep)
	case "glide":
		mgr = new(mgrGolangGlide)

	//node dependency managers
	case "npm":
		mgr = new(mgrNodeNpm)
	case "yarn":
		mgr = new(mgrNodeYarn)

	//python dependency managers
	case "pip":
		mgr = new(mgrPythonPip)

	//ruby dependency managers
	case "bundler":
		mgr = new(mgrRubyBundler)

	default:
		return nil, errors.ScmUnspecifiedError(fmt.Sprintf("Unknown Packager Manager Type: %s", mgrType))
	}

	if err := mgr.Init(pipelineData, config, client); err != nil {
		return nil, err
	}
	return mgr, nil
}

func Detect(packageType string, pipelineData *pipeline.Data, config config.Interface, client *http.Client) (Interface, error) {

	var mgrType string
	mgrType = "unknown"

	switch packageType {
	//chef dependency managers
	case "chef":
		if DetectChefBerkshelf(pipelineData, config, client) {
			mgrType = "berkshelf"
		} else { //default
			mgrType = "berkshelf"
		}

	//golang dependency managers
	case "golang":
		if DetectGolangDep(pipelineData, config, client) {
			mgrType = "dep"
		} else if DetectGolangGlide(pipelineData, config, client) {
			mgrType = "glide"
		} else { //default
			mgrType = "dep"
		}

	//node dependency managers
	case "node":
		if DetectNodeNpm(pipelineData, config, client) {
			mgrType = "npm"
		} else if DetectNodeYarn(pipelineData, config, client) {
			mgrType = "yarn"
		} else { //default
			mgrType = "npm"
		}

	//python dependency managers
	case "python":
		if DetectPythonPip(pipelineData, config, client) {
			mgrType = "pip"
		} else { //default
			mgrType = "pip"
		}

	//ruby dependency managers
	case "ruby":
		if DetectRubyBundler(pipelineData, config, client) {
			mgrType = "bundler"
		} else { //default
			mgrType = "bundler"
		}

	//empty/generic package manager. Noop.
	case "generic":
		mgrType = "generic"

	default:
		return nil, errors.MgrUnspecifiedError(fmt.Sprintf("Unknown Package Manager for Package Type (%s): %s", packageType, mgrType))
	}

	return Create(mgrType, pipelineData, config, client )
}
