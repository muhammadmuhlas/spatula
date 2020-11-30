package jira

import (
	"github.com/spf13/viper"
	"strings"
	"fmt"
)

type JQL struct {
	Query    string
}

func NewJQL(template string) (*JQL) {
	var q []string
	for _, t := range strings.Split(template, ",") {
		tpl := viper.GetString("pmt.template." + t)
		if strings.TrimSpace(tpl) != "" {
			q = append(q, tpl)
		}
	}
	return &JQL{Query: strings.Join(q, " AND ")}
}

func (jql *JQL) And(and string) *JQL {
	q := " AND "
	if strings.TrimSpace(jql.Query) == "" { q = "" }
	jql.Query = jql.Query + q + and
	return jql
}

func (jql *JQL) Or(or string) *JQL {
	q := " OR "
	if strings.TrimSpace(jql.Query) == "" { q = "" }
	jql.Query = jql.Query + q + or
	return jql
}

func (jql *JQL) Parse() string {
	fmt.Println("using query:", jql.Query)
	return strings.TrimSpace(jql.Query)
}
