package engine

import (
	"capsulecd/pkg/pipeline"
	"capsulecd/pkg/config"
	"capsulecd/pkg/scm"
)

type Interface interface {
	init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error
	ValidateTools() error
	BuildStep() error
	TestStep() error
	PackageStep() error
	DistStep() error
}

