package capsulecd

type Source interface {
	Configure()
	ProcessPushPayload()
	ProcessPullRequestPayload()
	Publish() //create release.
	Notify()
}