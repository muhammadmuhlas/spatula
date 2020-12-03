package jira

import (
	"github.com/andygrunwald/go-jira"
	"fmt"
	"github.com/spf13/viper"
	"github.com/parnurzeal/gorequest"
	"sync"
	"time"
)

type Issue struct {
	jira.Issue
	DevelopmentInfo DevelopmentInfo `json:"development_info"`
}

type DevelopmentInfo struct {
	Errors []interface{} `json:"errors"`
	Detail []struct {
		Branches []struct {
			Name                 string `json:"name"`
			URL                  string `json:"url"`
			CreatePullRequestURL string `json:"createPullRequestUrl"`
			Repository           struct {
				Name    string        `json:"name"`
				Avatar  string        `json:"avatar"`
				URL     string        `json:"url"`
				Commits []interface{} `json:"commits"`
			} `json:"repository"`
			LastCommit struct {
				ID              string        `json:"id"`
				DisplayID       string        `json:"displayId"`
				AuthorTimestamp string        `json:"authorTimestamp"`
				Merge           bool          `json:"merge"`
				Files           []interface{} `json:"files"`
			} `json:"lastCommit"`
		} `json:"branches"`
		PullRequests []struct {
			Author struct {
				Name   string `json:"name"`
				Avatar string `json:"avatar"`
			} `json:"author"`
			ID           string `json:"id"`
			Name         string `json:"name"`
			CommentCount int    `json:"commentCount"`
			Source       struct {
				Branch string `json:"branch"`
				URL    string `json:"url"`
			} `json:"source"`
			Destination struct {
				Branch string `json:"branch"`
				URL    string `json:"url"`
			} `json:"destination"`
			Reviewers  []interface{} `json:"reviewers"`
			Status     string        `json:"status"`
			URL        string        `json:"url"`
			LastUpdate string        `json:"lastUpdate"`
		} `json:"pullRequests"`
		Repositories []interface{} `json:"repositories"`
		Instance     struct {
			SingleInstance bool   `json:"singleInstance"`
			BaseURL        string `json:"baseUrl"`
			Name           string `json:"name"`
			TypeName       string `json:"typeName"`
			ID             string `json:"id"`
			Type           string `json:"type"`
		} `json:"_instance"`
	} `json:"detail"`
}

func (i *Issue) GetDevelopmentInfo(wg *sync.WaitGroup) (err error) {
	if wg != nil {
		defer wg.Done()
	}
	apiEndPoint := fmt.Sprintf("%s/rest/dev-status/latest/issue/detail?issueId=%s&applicationType=bitbucket&dataType=pullrequest", viper.GetString("pmt.credential.host"), i.ID)
	request := gorequest.New().Timeout(60*time.Second).SetBasicAuth(viper.GetString("pmt.credential.username"), viper.GetString("pmt.credential.password"))
	_, _, errs := request.Get(apiEndPoint).EndStruct(&i.DevelopmentInfo)
	if len(errs) > 0 {
		return errs[0]
	}
	return
}
