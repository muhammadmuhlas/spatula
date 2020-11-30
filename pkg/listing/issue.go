package listing

import (
	"github.com/andygrunwald/go-jira"
	"github.com/alexeyco/simpletable"
	"github.com/spf13/viper"
	"fmt"
	"strings"
	"strconv"
	"html/template"
	"bytes"
	"time"
	"errors"
	"github.com/muhammadmuhlas/spatula/pkg/formatter"
)

type (
	// Issue defines the properties of an issue to be listed
	Issue struct {
		jira.Issue
		DevelopmentInfo DevelopmentInfo `json:"development_info"`
	}
	DevelopmentInfo struct {
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
		} `json:"detail"`
	}
	Lite struct {
		IssueKey string `json:"issue_key"`
		IssueURL string `json:"issue_url"`
		Summary  string `json:"summary"`
	}
	BranchLite struct {
		Owner      string `json:"owner"`
		Repository string `json:"repository"`
		Branch     string `json:"branch"`
	}
	Changelog struct {
		Date    string `json:"date"`
		Version string `json:"version"`
		Bugs    []Lite `json:"bugs"`
		Tasks   []Lite `json:"tasks"`
	}
	RepositoryChangelog struct {
		Repository string
		Changelog  Changelog
	}
)

//DisplayTasks display issues as table output
func DisplayTasks(kind string, data []Issue, version, date string) (string, error) {
	switch kind {
	case "table":
		return table(data), nil
	case "changelog":
		return changelog(data, false, version, date)
	case "release":
		return changelog(data, true, version, date)
	default:
		return "", errors.New("unsupported output type")
	}
}

func ToArray(data []Issue) (s []string) {
	for _, datum := range data {

		s = append(s, fmt.Sprintf("%s = %s (%s) %s", datum.Key, func() string {
			if datum.Fields.Assignee != nil {
				return formatter.Blue(formatter.Abbr(datum.Fields.Assignee.DisplayName))
			}
			return formatter.Blue("Unassigned")
		}(), formatter.Gray(datum.Fields.Status.Name), func() string {
			summ := strings.TrimSpace(datum.Fields.Summary)
			if len(summ) > 100 {
				return formatter.Red(datum.Fields.Summary[:100]) + "..."
			}
			return formatter.Red(summ)
		}()))
	}
	return
}

func ToIssue(s []string) (iss []string) {
	for _, is := range s {
		iss = append(iss, strings.TrimSpace(strings.Split(is, "=")[0]))
	}
	return iss
}

func table(data []Issue) string {
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "Issue Key"},
			{Align: simpletable.AlignCenter, Text: "SP"},
			{Align: simpletable.AlignLeft, Text: "Type"},
			{Align: simpletable.AlignLeft, Text: "Assignee"},
			{Align: simpletable.AlignCenter, Text: "Summary"},
			{Align: simpletable.AlignCenter, Text: "Status"},
			{Align: simpletable.AlignCenter, Text: "Repository"},
		},
	}
	var tmp [][]*simpletable.Cell
	for _, v := range data {
		tmp = append(tmp, []*simpletable.Cell{
			{Text: formatter.Blue(v.Key)},
			{Text: func() string {
				a, _ := v.Fields.Unknowns.Value(viper.GetString("pmt.fields.story-point"))
				if a != nil {
					return formatter.Blue(strconv.Itoa(int(a.(float64))))
				}
				return ""
			}()},
			{Text: func() string {
				tp := strings.ToLower(v.Fields.Type.Name)
				if tp == "epic" {
					return formatter.Blue(v.Fields.Type.Name)
				}
				if tp == "task" {
					return formatter.Gray(v.Fields.Type.Name)
				}
				if tp == "bug" {
					return formatter.Red(v.Fields.Type.Name)
				}
				return v.Fields.Type.Name
			}()},
			{Text: func() string {
				if v.Fields.Assignee != nil {
					return formatter.Abbr(v.Fields.Assignee.DisplayName)
				}
				return "Unassigned"
			}()},
			{Text: func() string {
				summ := v.Fields.Summary
				if len(summ) > 100 {
					return strings.TrimSpace(v.Fields.Summary[:100]) + "..."
				}
				return strings.TrimSpace(v.Fields.Summary)
			}()},
			{Text: func() string {
				status := strings.ToLower(v.Fields.Status.Name)
				if status == "done" || status == "resolved" {
					return formatter.Green(v.Fields.Status.Name)
				}
				return formatter.Red(v.Fields.Status.Name)
			}()},
			{Text: func() string {
				var s []string
				for _, b := range repository(v.DevelopmentInfo) {
					s = append(s, fmt.Sprintf("%s@%s", b.Repository, b.Branch))
				}
				return strings.Join(s, ", ")
			}()},
		})
	}
	table.Body = &simpletable.Body{
		Cells: tmp,
	}
	table.SetStyle(simpletable.StyleUnicode)
	return table.String()
}

func changelog(data []Issue, grouped bool, version, date string) (string, error) {
	tpl := template.Must(
		template.New("changelog.notes").
			Funcs(template.FuncMap{"trim": strings.TrimSpace}).
			Parse(viper.GetString("scm.template.changelog.notes")))
	var buff bytes.Buffer
	if grouped {
		rch, err := issuesToRepositoryChangelog(data, version, date)
		if err != nil {
			return "", err
		}
		output := fmt.Sprintf("Below Contains %d Changelogs\n=================\n\n", len(rch))
		for i, repositoryChangelog := range rch {
			output += fmt.Sprintf("%d. %s\n", i+1, repositoryChangelog.Repository)
			if err := tpl.Execute(&buff, repositoryChangelog.Changelog); err != nil {
				return "", err
			}
			output += strings.TrimSpace(buff.String())
			buff.Reset()
			output += fmt.Sprintf("\n=============\n\n")
		}
		output += fmt.Sprintf("=================\nEnd Of Changelog\n")
		return output, err
	}
	ch, err := issuesToChangelog(data, version, date)
	if err != nil {
		return "", err
	}
	if err := tpl.Execute(&buff, ch); err != nil {
		return "", err
	}
	return buff.String(), err
}

func issuesToChangelog(iss []Issue, version, date string) (ch Changelog, err error) {
	for _, issue := range iss {
		if strings.ToLower(issue.Fields.Type.Name) == "bug" {
			ch.Bugs = append(ch.Bugs, Lite{
				IssueKey: issue.Key,
				IssueURL: viper.GetString("pmt.credential.host") + "/browse/" + issue.Key,
				Summary:  issue.Fields.Summary,
			})
			continue
		}
		if strings.ToLower(issue.Fields.Type.Name) == "task" {
			ch.Tasks = append(ch.Tasks, Lite{
				IssueKey: issue.Key,
				IssueURL: viper.GetString("pmt.credential.host") + "/browse/" + issue.Key,
				Summary:  strings.TrimSpace(issue.Fields.Summary),
			})
			continue
		}
	}
	if date == "" {
		date = time.Now().Format("02 January 2006")
	}
	ch.Date = date
	ch.Version = version
	return
}

func issuesToRepositoryChangelog(iss []Issue, version, date string) (rch []RepositoryChangelog, err error) {
	rissue := make(map[string][]Issue)
	for _, issue := range iss {
		for _, dd := range issue.DevelopmentInfo.Detail {
			for _, ddb := range dd.Branches {
				if rissue[ddb.Repository.Name] == nil {
					rissue[ddb.Repository.Name] = []Issue{}
				}
				rissue[ddb.Repository.Name] = append(rissue[ddb.Repository.Name], issue)
			}
		}
	}
	for s, issues := range rissue {
		ch, err := issuesToChangelog(issues, version, date)
		if err != nil {
			fmt.Println(err)
		}
		rch = append(rch, RepositoryChangelog{
			Repository: s,
			Changelog:  ch,
		})
	}
	return
}

func repository(infos DevelopmentInfo) (result []BranchLite) {
	for _, details := range infos.Detail {
		for _, branch := range details.Branches {
			result = append(result, BranchLite{
				Repository: branch.Repository.Name,
				Branch:     branch.Name,
			})
		}
	}
	return
}
