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
	"github.com/spf13/viper"
	"github.com/muhammadmuhlas/spatula/pkg/formatter"
	"fmt"
)

// initConfigCmd represents the initConfig command
var initConfigCmd = &cobra.Command{
	Use:   "init",
	Short: "Write initial config files",
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("scm.provider", "bitbucket")
		viper.SetDefault("scm.credential.workspace", "user")
		viper.SetDefault("scm.credential.username", "email@example.com")
		viper.SetDefault("scm.credential.password", "token")
		viper.SetDefault("scm.template.pull-request.title", "{{ .Source.Branch }} to {{ .Destination.Branch }}")
		viper.SetDefault("scm.template.changelog.commit", "Prepare release for {{ .version }}")
		viper.SetDefault("scm.template.changelog.notes", `---
# {{ .Version }}
Released {{ .Date }}

{{- if .Tasks }}
## Tasks

{{- end }}
{{- range .Tasks }}
*   [[{{ .IssueKey }}]({{ .IssueURL }})] - {{ .Summary }}
{{- end }}

{{- if .Bugs }}
## Bug

{{- end }}
{{- range .Bugs }}
*   [[{{ .IssueKey }}]({{ .IssueURL }})] - {{ .Summary }}
{{- end }}

---`)
		viper.SetDefault("pmt.provider", "jira")
		viper.SetDefault("pmt.credential.host", "https://company.atlassian.net")
		viper.SetDefault("pmt.credential.username", "email@example.com")
		viper.SetDefault("pmt.credential.password", "token")
		viper.SetDefault("pmt.fields.story-point", "customfield_10016")
		viper.SetDefault("pmt.template.default", "")
		viper.SetDefault("pmt.template.mine", "assignee in (currentUser())")
		viper.SetDefault("pmt.template.version", "fixVersion")
		viper.SetConfigType("yaml")

		pmtProv := formatter.PromptQuestion("Enter Your PMT Provider", "jira")
		viper.Set("pmt.provider", pmtProv)

		if pmtProv == "jira" {
			fmt.Println("create a new one on https://id.atlassian.com/manage-profile/security/api-tokens")
		}

		pmtHost := formatter.PromptQuestion("[PMT] Enter Your PMT Host", "https://company.atlassian.net")
		viper.Set("pmt.credential.host", pmtHost)

		pmtUser := formatter.PromptQuestion("[PMT] Enter Your PMT Username", "user@example.com")
		viper.Set("pmt.credential.username", pmtUser)

		pmtPass := formatter.PasswordQuestion("[PMT] Enter Your PMT Password")
		viper.Set("pmt.credential.password", pmtPass)

		scmProv := formatter.PromptQuestion("Enter Your SCM Provider", "bitbucket")
		viper.Set("scm.provider", scmProv)

		if scmProv == "bitbucket" {
			fmt.Println("create a new one on https://bitbucket.org/account/settings/app-passwords/")
		}

		scmUser := formatter.PromptQuestion("[SCM] Enter Your SCM Username", "awesome_user")
		viper.Set("scm.credential.username", scmUser)

		scmPass := formatter.PasswordQuestion("[SCM] Enter Your SCM Password")
		viper.Set("scm.credential.password", scmPass)

		scmHost := formatter.PromptQuestion("[SCM] Enter Your SCM Workspace", "user")
		viper.Set("scm.credential.workspace", scmHost)

		viper.SafeWriteConfigAs("./.spatula.yaml")
	},
}

func init() {
	rootCmd.AddCommand(initConfigCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initConfigCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initConfigCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
