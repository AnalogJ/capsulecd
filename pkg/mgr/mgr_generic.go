package mgr

import (
	"capsulecd/pkg/pipeline"
	"net/http"
	"capsulecd/pkg/config"
)

func DetectGeneric(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	return false
}


type mgrGeneric struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}


func (m *mgrGeneric) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrGeneric) MgrValidateTools() error {
	return nil
}

func (m *mgrGeneric) MgrAssembleStep() error {
	return nil
}

func (m *mgrGeneric) MgrDependenciesStep(currentMetadata interface{}, nextMetadata interface{}) error {
	return nil
}

func (m *mgrGeneric) MgrPackageStep(currentMetadata interface{}, nextMetadata interface{}) error {
	return nil
}


func (m *mgrGeneric) MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error {
	return nil
}