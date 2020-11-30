package working

import (
	"github.com/muhammadmuhlas/spatula/pkg/listing"
	"github.com/AlecAivazis/survey/v2"
	"fmt"
	"strings"
	"github.com/muhammadmuhlas/spatula/pkg/formatter"
	"github.com/AlecAivazis/survey/v2/terminal"
	"os"
	"github.com/spf13/viper"
)

const Release = "Release"
const Merge = "Merge"

func ChooseWorkMode() (ans string) {
	prompt := &survey.Select{
		Message: "Please Choose Work Mode",
		Options: []string{Release, Merge},
	}
	err := survey.AskOne(prompt, &ans, survey.WithValidator(survey.Required))
	if err == terminal.InterruptErr {
		fmt.Println("interrupted")

		os.Exit(0)
	} else if err != nil {
		panic(err)
	}
	return
}

func ChooseReleaseVersion(version string) (ans string) {
	prompt := &survey.Input{
		Message: "Enter Release Version",
		Default: version,
	}
	err := survey.AskOne(prompt, &ans, survey.WithValidator(survey.Required))
	if err == terminal.InterruptErr {
		fmt.Println("interrupted")

		os.Exit(0)
	} else if err != nil {
		panic(err)
	}
	return strings.TrimSpace(ans)
}

func ChooseDestBranch(kind, version string) (ans string) {
	prompt := &survey.Input{
		Message: "Enter " + kind + " Branch",
	}
	err := survey.AskOne(prompt, &ans, survey.WithValidator(survey.Required))
	if err == terminal.InterruptErr {
		fmt.Println("interrupted")

		os.Exit(0)
	} else if err != nil {
		panic(err)
	}
	if version == "" {
		return strings.TrimSpace(ans)
	}
	return strings.TrimSpace(ans) + "/" + version
}

func ChooseIssueKeys(iss []listing.Issue, branch string) (ans []string) {
	prompt := &survey.MultiSelect{
		Message: "Please Choose Related Issues For " + branch,
		Options: listing.ToArray(iss),
	}
	err := survey.AskOne(prompt, &ans, survey.WithValidator(survey.Required))
	if err == terminal.InterruptErr {
		fmt.Println("interrupted")

		os.Exit(0)
	} else if err != nil {
		panic(err)
	}
	return listing.ToIssue(ans)
}

func ChooseIssueBranches(iss []listing.Issue, releaseBranch string) (bl PullRequestQueue) {
	ans := []string{}
	prompt := &survey.MultiSelect{
		Message: "Please Choose Branches Of Issue Keys To Open PR",
		Options: ToArray(iss, releaseBranch),
	}
	err := survey.AskOne(prompt, &ans, survey.WithValidator(survey.Required), survey.WithPageSize(15))
	if err == terminal.InterruptErr {
		fmt.Println("interrupted")

		os.Exit(0)
	} else if err != nil {
		panic(err)
	}
	return ToBranchLite(ans)
}

func ConfirmRelease(question string) (ans bool) {
	prompt := &survey.Confirm{
		Message: question,
		Default: true,
	}
	err := survey.AskOne(prompt, &ans, survey.WithValidator(survey.Required))
	if err == terminal.InterruptErr {
		fmt.Println("interrupted")

		os.Exit(0)
	} else if err != nil {
		panic(err)
	}
	return
}

func ToArray(data []listing.Issue, releaseBranch string) (s []string) {
	for _, datum := range data {
		for _, devDetail := range datum.DevelopmentInfo.Detail {
			for _, branch := range devDetail.Branches {
				s = append(s, fmt.Sprintf("[%s] %s@%s => %s@%s", formatter.Red(datum.Key), branch.Repository.Name, branch.Name, formatter.Blue(branch.Repository.Name), formatter.Blue(releaseBranch)))
			}
		}
	}
	if len(s) < 1 { panic("No Development Info!")}
	return
}

func ToBranchLite(lists []string) (bl PullRequestQueue) {
	for _, is := range lists {
		var source, target listing.BranchLite
		for i, slist := range strings.Split(strings.TrimSpace(formatter.Decolorize(is)), " ") {
			if i == 1 {
				source.Repository = strings.Split(slist, "@")[0]
				source.Branch = strings.Split(slist, "@")[1]
				source.Owner = viper.GetString("scm.credential.workspace")
			}
			if i == 3 {
				target.Repository = strings.Split(slist, "@")[0]
				target.Branch = strings.Split(slist, "@")[1]
				target.Owner = viper.GetString("scm.credential.workspace")
			}
		}
		pr := PullRequestTask{
			Source:      source,
			Destination: target,
		}
		title, err := formatter.TemplateExecutor("scm.template.pull-request.title", pr)
		if err != nil { return nil }
		pr.Title = title
		bl = append(bl, &pr)
	}
	return bl
}
