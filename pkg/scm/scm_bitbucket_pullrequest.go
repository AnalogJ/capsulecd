package scm

import (
	"time"
)

type scmBitbucketPullrequest struct {
	CreatedOn time.Time `json:"created_on"`
	PullRequestNumber string `json:"id"`
	State string `json:"state"`
	Title string `json:"title"`
	Base struct {
		Branch struct {
			Name string  `json:"name"`
		}  `json:"branch"`
		Commit struct {
			Hash string `json:"string"`
	       	} `json:"commit"`
		Repository struct {
			FullName string `json:"full_name"`
			Name string `json:"name"`
		} `json:"repository"`
	}  `json:"destination"`

	Head struct {
	     Branch struct {
			    Name string  `json:"name"`
		    }  `json:"branch"`
	     Commit struct {
			    Hash string `json:"string"`
		    } `json:"commit"`
	     Repository struct {
			    FullName string `json:"full_name"`
			    Name string `json:"name"`
		    } `json:"repository"`
     	}  `json:"source"`

}

