package engine

import (
	"capsulecd/lib/config"
	"capsulecd/lib/errors"
	"fmt"
	"capsulecd/lib/scm"
)

type Engine interface {
	ValidateTools() error
	Init(sourceScm * scm.Scm) error
	BuildStep() error
	TestStep() error
	PackageStep() error
	ReleaseStep() error
}

func Create() (Engine, error) {

	switch engineType := config.Get("package_type"); engineType {
	case "chef":
		return &engineChef{}, nil
	case "golang":
		return &engineChef{}, nil
	case "javascript":
		return &engineChef{}, nil
	case "node":
		return &engineChef{}, nil
	case "python":
		return &engineChef{}, nil
	case "ruby":
		return &engineChef{}, nil
	default:
		return nil, errors.EngineUnspecifiedError(fmt.Sprintf("Unknown Engine Type: %s", engineType))
	}
}