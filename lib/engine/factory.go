package engine

import (
	"capsulecd/lib/config"
	"capsulecd/lib/errors"
	"capsulecd/lib/pipeline"
	"capsulecd/lib/scm"
	"fmt"
)

type Engine interface {
	ValidateTools() error
	Init(pipelineData *pipeline.Data, sourceScm scm.Scm) error
	BuildStep() error
	TestStep() error
	PackageStep() error
	DistStep() error
}

func Create() (Engine, error) {

	switch engineType := config.Get("package_type"); engineType {
	case "chef":
		return new(engineChef), nil
	case "golang":
		return new(engineChef), nil
	case "node":
		return new(engineNode), nil
	case "python":
		return new(enginePython), nil
	case "ruby":
		return new(engineRuby), nil
	default:
		return nil, errors.EngineUnspecifiedError(fmt.Sprintf("Unknown Engine Type: %s", engineType))
	}
}
