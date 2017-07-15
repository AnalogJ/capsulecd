package scm

import (
	"capsulecd/lib/config"
	"log"
)


type Scm interface {
	Options() *ScmOptions
	Configure()
	RetrievePayload() *ScmPayload
	ProcessPushPayload()
	ProcessPullRequestPayload()
	Publish() //create release.
	Notify()
}

func Create() Scm {

	switch scmType := config.Get("scm"); scmType {
	case "bitbucket":
		return &scmBitbucket{}
	case "github":
		return &scmGithub{}
	default:
		log.Fatal("Unknown scm type")
		return nil
	}
}