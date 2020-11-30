package scm

import (
	"errors"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/muhammadmuhlas/spatula/pkg/working"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"net/http"
	"github.com/parnurzeal/gorequest"
	"time"
	"fmt"
	"github.com/muhammadmuhlas/spatula/pkg/formatter"
)

type SCM struct {
	scm *bitbucket.Client
}

func NewProvider(provider, username, token string) (*SCM, error) {
	switch provider {
	case "bitbucket":
		bb := bitbucket.NewBasicAuth(username, token)
		return &SCM{scm: bb}, nil
	default:
		return nil, errors.New("unrecognized provider")
	}
}

func (p *SCM) CreatePullRequest(q working.PullRequestTask) (res interface{}, err error) {
	var opt bitbucket.PullRequestsOptions
	opt.Owner = q.Destination.Owner
	opt.RepoSlug = q.Destination.Repository
	opt.Title = q.Title
	opt.SourceBranch = q.Source.Branch
	opt.DestinationBranch = q.Destination.Branch
	opt.Description = formatter.DefaultPullRequestDescription()
	return p.scm.Repositories.PullRequests.Create(&opt)
}

func (p *SCM) CreateBranch(owner, repository, name, source string) (rb working.Branch, err error) {
	payload := make(map[string]interface{})
	payload["name"] = name
	payload["target"] = map[string]interface{}{}
	payload["target"].(map[string]interface{})["hash"] = source
	apiEndPoint := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/refs/branches", owner, repository)
	request := gorequest.New().Retry(5, 5 * time.Second, http.StatusTooManyRequests, http.StatusBadGateway).SetBasicAuth(viper.GetString("scm.credential.username"), viper.GetString("scm.credential.password"))
	_, _, errs := request.Post(apiEndPoint).Send(payload).EndStruct(&rb)
	if len(errs) > 0 { panic(errs)}
	return
}

func (p *SCM) GetBranch(owner, repository, name string) (rb working.Branch, err error) {
	var brb bitbucket.RepositoryBranchOptions
	brb.BranchName = name
	brb.Owner = owner
	brb.RepoSlug = repository
	res, err := p.scm.Repositories.Repository.GetBranch(&brb)
	if err != nil {
		return
	}
	if err := copier.Copy(&rb, &res); err != nil {
		return rb, err
	}
	return
}
