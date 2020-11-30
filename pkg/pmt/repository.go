package pmt

import (
	"github.com/andygrunwald/go-jira"
	"github.com/muhammadmuhlas/spatula/pkg/listing"
	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"github.com/briandowns/spinner"
	"time"
	"fmt"
	"sort"
	jira2 "github.com/muhammadmuhlas/spatula/pkg/pmt/jira"
	"errors"
)

type PMT struct {
	pmt *jira.Client
}

func NewProvider(provider, username, token string) (*PMT, error){
	switch provider {
	case "jira":
		tp := jira.BasicAuthTransport{
			Username: username,
			Password: token,
		}
		c, err := jira.NewClient(tp.Client(), viper.GetString("pmt.credential.host"))
		if err != nil {
			return nil, err
		}
		return &PMT{pmt: c}, nil
	default:
		return nil, errors.New("unrecognized provider")
	}
}

func (p *PMT) GetIssue(id string) (is listing.Issue, err error) {
	var i jira2.Issue
	issues, _, err := p.pmt.Issue.Get(id, nil)
	if err := copier.Copy(&i, &issues); err != nil {
		return listing.Issue{}, err
	}
	if err := i.GetDevelopmentInfo(nil); err != nil {
		return is, err
	}
	if err := copier.Copy(&is, &i); err != nil {
		return listing.Issue{}, err
	}
	return is, err
}

func (p *PMT) GetAllIssues(jql string, size int, dev bool) (iss []listing.Issue, err error) {
	var pmtIssues []*jira2.Issue
	start := time.Now()
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()
	s.Color("red", "bold")
	def := 50
	if size < 50 {
		def = size
	}

	for i := 0; i < size; i += def {
		s.Prefix = fmt.Sprintf("Fetching %d of %d (%.2fs elapsed) ", i, size, time.Since(start).Seconds())
		opts := jira.SearchOptions{MaxResults: def, StartAt: i}
		issues, _, _ := p.pmt.Issue.Search(jql, &opts)
		if err := copier.Copy(&pmtIssues, &issues); err != nil {
			return nil, err
		}
	}
	sort.Slice(pmtIssues, func(i, j int) bool {
		return pmtIssues[i].Key > pmtIssues[j].Key
	})

	if dev {
		//var wg sync.WaitGroup
		for i, pmtIssue := range pmtIssues {
			s.Prefix = fmt.Sprintf("[%d/%d] Getting development info for %s, it may take longer due to rate limit policy...", i+1, len(pmtIssues), pmtIssue.Key)
			//wg.Add(1)
			if i%5 == 0 {
				time.Sleep(3 * time.Second)
			}
			pmtIssue.GetDevelopmentInfo(nil)
		}
		//wg.Wait()
	}

	if err := copier.Copy(&iss, &pmtIssues); err != nil {
		return nil, err
	}

	s.Stop()
	return
}
