/*
Copyright © 2020 Muhammad Muhlas Abror <muhammadmuhlas3@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"github.com/muhammadmuhlas/spatula/pkg/listing"
	"github.com/muhammadmuhlas/spatula/pkg/pmt/jira"
	"strings"
	"github.com/spf13/viper"
	"time"
	"github.com/muhammadmuhlas/spatula/pkg/working"
	"github.com/briandowns/spinner"
	"errors"
	"github.com/muhammadmuhlas/spatula/pkg/pmt"
	"github.com/muhammadmuhlas/spatula/pkg/scm"
	"github.com/muhammadmuhlas/spatula/pkg/formatter"
)

// doCmd represents the release command
var doCmd = &cobra.Command{
	Use:   "do",
	Short: "Do some works (release, merge, generate release notes)",
	RunE: func(cmd *cobra.Command, args []string) error {

		pmtProvider, err := pmt.NewProvider(viper.GetString("pmt.provider"), viper.GetString("pmt.credential.username"), viper.GetString("pmt.credential.password"))
		if err != nil {
			return err
		}
		scmProvider, err := scm.NewProvider(viper.GetString("scm.provider"), viper.GetString("scm.credential.username"), viper.GetString("scm.credential.password"))
		if err != nil {
			return err
		}
		pmt := listing.NewService(pmtProvider)
		scm := working.NewService(scmProvider)

		var iss []listing.Issue
		template, err := cmd.Flags().GetString("template")
		project, err := cmd.Flags().GetString("project")
		date, err := cmd.Flags().GetString("date")
		size, err := cmd.Flags().GetInt("size")

		if len(args) > 0 {
			tmp := strings.Join(args, " ")
			for _, v := range strings.Split(tmp, ",") {
				issue, err := pmt.GetIssue(v)
				if err != nil {
					return err
				}
				iss = append(iss, issue)
			}
		}

		fmt.Println("Using template:", template)
		tpl := jira.NewJQL(template)
		if project != "" {
			tpl.And(`project = "` + project + `"`)
		}
		issues, err := pmtProvider.GetAllIssues(tpl.Parse(), size, false)
		if err != nil {
			return err
		}
		for _, issue := range issues {
			iss = append(iss, issue)
		}
		fmt.Printf("Found: %d issues\n\n", len(issues))
		if len(issues) < 1 {
			return errors.New("No Issues can be released using this query!")
		}

		workMode := working.ChooseWorkMode()
		fversion, err := cmd.Flags().GetString("version")
		var version,targetBranch string
		if workMode == working.Release {
			version = working.ChooseReleaseVersion(fversion)
		}

		targetBranch = working.ChooseDestBranch(workMode, version)

		issueKeys := working.ChooseIssueKeys(iss, targetBranch)

		start := time.Now()
		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Start()
		s.Color("red", "bold")
		var chosen []listing.Issue
		for i, v := range issueKeys {
			s.Prefix = fmt.Sprintf("Obtaining development info %d of %d (%.2fs elapsed) ", i, len(issueKeys), time.Since(start).Seconds())
			issue, err := pmt.GetIssue(v)
			if err != nil {
				return err
			}
			chosen = append(chosen, issue)
		}
		s.Stop()

		prQueue := working.ChooseIssueBranches(chosen, targetBranch)

		if !working.ConfirmRelease("Next Step Is Creating PR, Are You Sure Want to Continue?") {
			return errors.New("release not confirmed. cancelling the rest")
		}
		for _, task := range prQueue {
			s.Prefix = "Execute Pull Request Queue"
			s.Start()

			s.Prefix = fmt.Sprintf("Checking Source Branch %s", formatter.Red(task.Source.Branch))
			if !scm.BranchExist(task.Source.Owner, task.Source.Repository, task.Source.Branch) {
				s.Prefix = fmt.Sprintf("[PR] Creating Branch %s", formatter.Red(task.Source.Branch))
				if _, err := scm.CreateBranch(task.Source.Owner, task.Source.Repository, task.Source.Branch, "master"); err != nil {
					return err
				}
				s.Stop()
				fmt.Println("New Branch", task.Source.Branch, "Created!")
				s.Start()
			}
			s.Prefix = fmt.Sprintf("Checking Destination Branch %s", formatter.Red(task.Destination.Branch))
			if !scm.BranchExist(task.Destination.Owner, task.Destination.Repository, task.Destination.Branch) {
				if _, err := scm.CreateBranch(task.Destination.Owner, task.Source.Repository, task.Destination.Branch, "master"); err != nil {
					return err
				}
				s.Stop()
				fmt.Println("New Branch", task.Destination.Branch, "Created!")
				s.Start()
			}

			s.Prefix = fmt.Sprintf("Creating Pull Request On %s \"%s\" ", formatter.Red(task.Destination.Repository), formatter.Blue(task.Title))
			scm, err := scm.CreatePullRequest(*task)
			if err != nil {
				s.Stop()
				fmt.Printf("Pull Request Can't Be Opened!. No Diff %s - %s", formatter.Red(task.Destination.Repository), formatter.Red(task.Source.Repository))
				continue
			}
			s.Stop()
			fmt.Println(fmt.Sprintf("Pull Request #%d Opened On %s \"%s\" %s", int(scm.(map[string]interface{})["id"].(float64)), formatter.Red(task.Destination.Repository), formatter.Blue(task.Title), scm.(map[string]interface{})["links"].(map[string]interface{})["html"].(map[string]interface{})["href"].(string)))
		}

		if working.ConfirmRelease("Do You Need a Generated Changelog For this Release?") {
			rNote, err := listing.DisplayTasks("release", chosen, version, date)
			if err != nil { return err
			}
			fmt.Println(rNote)
		}

		if working.ConfirmRelease("Do You Need Pull Request From " + targetBranch + " To Main Branch?") {
			mainBranch := formatter.PromptQuestion("What Is Your Main Branch?", "master")
			if working.ConfirmRelease("Next Step Is Creating PR, Are You Sure Want to Continue?") {
				for _, task := range prQueue {
					dest := task.Destination
					task.Source = dest
					task.Destination = listing.BranchLite{
						Repository: dest.Repository,
						Branch:     mainBranch,
						Owner: viper.GetString("scm.credential.workspace"),
					}
					task.Title = task.Source.Branch + " to " + task.Destination.Branch
					// DEBUG
					task.Destination.Repository = "a"

					s.Prefix = fmt.Sprintf("Creating Pull Request On %s \"%s\" ", formatter.Red(task.Destination.Repository), formatter.Blue(task.Title))
					s.Start()
					scm, err := scm.CreatePullRequest(*task)
					if err != nil {
						s.Stop()
						fmt.Println(fmt.Sprintf("Pull Request Can't Be Opened!. No Diff %s - %s", formatter.Red(task.Destination.Branch), formatter.Red(task.Source.Branch)))
						panic(err)
					}
					s.Stop()
					fmt.Println(fmt.Sprintf("Pull Request #%d Opened On %s \"%s\" %s", int(scm.(map[string]interface{})["id"].(float64)), formatter.Red(task.Destination.Repository), formatter.Blue(task.Title), scm.(map[string]interface{})["links"].(map[string]interface{})["html"].(map[string]interface{})["href"].(string)))
				}
			}
		}

		if working.ConfirmRelease("Thank you!") {
			fmt.Println("You're Welcome")
			return nil
		}


		return nil
	},
}

func init() {
	rootCmd.AddCommand(doCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// doCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// doCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	doCmd.PersistentFlags().String("template", "default", "Template for searching issues")
	doCmd.PersistentFlags().String("project", "", "add project prefix to issues")
	doCmd.PersistentFlags().String("version", "", "release version")
	doCmd.PersistentFlags().String("date", time.Now().Format("02 January 2006"), "date for changelog output")
	doCmd.PersistentFlags().Int("size", 50, "size for listing issues")
}
