package working

import "github.com/muhammadmuhlas/spatula/pkg/listing"

type PullRequestTask struct {
	Title       string
	Source      listing.BranchLite
	Destination listing.BranchLite
}

type PullRequestQueue []*PullRequestTask