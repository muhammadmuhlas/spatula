package bitbucket

import "github.com/ktrysmt/go-bitbucket"

type Branch struct {
	bitbucket.RepositoryBranchOptions
	bitbucket.RepositoryBranch
}
