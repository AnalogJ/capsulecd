package pipeline

import "time"

type GitTagDetails struct {
	TagShortName string
	CommitSha    string
	CommitDate   time.Time
}
