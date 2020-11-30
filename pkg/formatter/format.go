package formatter

import (
	"fmt"
	"strings"
	"bytes"
	"github.com/spf13/viper"
	"html/template"
	"os/user"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"os"
)

const (
	ColorDefault = "\x1b[39m"

	ColorRed   = "\x1b[91m"
	ColorGreen = "\x1b[32m"
	ColorBlue  = "\x1b[94m"
	ColorGray  = "\x1b[90m"
)

func Red(s string) string {
	return fmt.Sprintf("%s%s%s", ColorRed, s, ColorDefault)
}

func Green(s string) string {
	return fmt.Sprintf("%s%s%s", ColorGreen, s, ColorDefault)
}

func Blue(s string) string {
	return fmt.Sprintf("%s%s%s", ColorBlue, s, ColorDefault)
}

func Gray(s string) string {
	return fmt.Sprintf("%s%s%s", ColorGray, s, ColorDefault)
}

func Decolorize(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, ColorGray, ""), ColorBlue, ""), ColorGreen, ""), ColorRed, ""), ColorDefault, "")
}

func Abbr(s string) (o string) {
	ss := strings.Split(strings.TrimSpace(s), " ")
	for i, str := range ss {
		if i > 0 {
			o += str[:1]
			continue
		}
		o += str + " "
	}
	return strings.TrimSpace(o)
}

func TemplateExecutor(configTemplate string, data interface{}) (string, error) {
	tpl := template.Must(
		template.New("template").
			Funcs(template.FuncMap{"trim": strings.TrimSpace}).
			Parse(viper.GetString(configTemplate)))
	var buff bytes.Buffer
	if err := tpl.Execute(&buff, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(buff.String()), nil
}

func CurrentUser() string {
	user, err := user.Current()
	if err != nil {
		return "anonymous"
	}
	return user.Username
}

func DefaultPullRequestDescription() string {
	return fmt.Sprintf("\n#### Notes\n> This Pull Request Opened By **%s** Using [**Spatula**](https://github.com/muhammadmuhlas/spatula)", CurrentUser())
}

func PromptQuestion(question, def string) (ans string) {
	prompt := &survey.Input{
		Message: question,
		Default: def,
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

func PasswordQuestion(question string) (ans string) {
	prompt := &survey.Password{
		Message: question,
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

