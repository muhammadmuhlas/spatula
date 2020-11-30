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
	"github.com/muhammadmuhlas/spatula/pkg/pmt/jira"
	"fmt"
	"github.com/spf13/viper"
	"github.com/muhammadmuhlas/spatula/pkg/listing"
	"strings"
	"time"
	"github.com/muhammadmuhlas/spatula/pkg/pmt"
)

// issueCmd represents the issue command
var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage issues of specific projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		var iss []listing.Issue
		provider, err := pmt.NewProvider(viper.GetString("pmt.provider"), viper.GetString("pmt.credential.username"), viper.GetString("pmt.credential.password"))
		if err != nil {
			return err
		}
		service := listing.NewService(provider)

		template, err := cmd.Flags().GetString("template")
		project, err := cmd.Flags().GetString("project")
		output, err := cmd.Flags().GetString("output")
		version, err := cmd.Flags().GetString("version")
		date, err := cmd.Flags().GetString("date")
		size, err := cmd.Flags().GetInt("size")
		dev, err := cmd.Flags().GetBool("with-dev")

		if len(args) > 0 {
			tmp := strings.Join(args, " ")
			for _, v := range strings.Split(tmp, ",") {
				issue, err := service.GetIssue(v)
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
		issues, err := provider.GetAllIssues(tpl.Parse(), size, dev)
		if err != nil {
			return err
		}
		for _, issue := range issues { iss = append(iss, issue) }
		fmt.Printf("Found: %d issues\n\n", len(issues))

		fmt.Println("Using output:", output)
		out, err := listing.DisplayTasks(output, iss, version, date)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return err
	},
}

func init() {
	rootCmd.AddCommand(issueCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// issueCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	issueCmd.PersistentFlags().String("template", "default", "Template for searching issues")
	issueCmd.PersistentFlags().String("project", "", "add project prefix to issues")
	issueCmd.PersistentFlags().String("output", "table", "ouput type (table or changelog)")
	issueCmd.PersistentFlags().Bool("with-dev", false, "include development info")
	issueCmd.PersistentFlags().String("version", "1.0.0", "version for changelog output")
	issueCmd.PersistentFlags().String("date", time.Now().Format("02 January 2006"), "date for changelog output")
	issueCmd.PersistentFlags().Int("size", 50, "size for listing issues")
}
