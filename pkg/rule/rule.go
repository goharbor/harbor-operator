package rule

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type Rule struct {
	RegistryRegex string
	Project       string
	ServerURL     string
}

// StringToRules parse rule and create Rule object
// assume rule.Rules are concatentated by '=>'.
func StringToRules(raw []string, server string) ([]Rule, error) {
	res := make([]Rule, 0)

	// remove https/http from the serverURL
	u, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	for _, r := range raw {
		// format read from configMap could be like '- docker.io=>value'
		if r == "" || !strings.Contains(r, "=>") {
			return nil, errors.Errorf("rule '%s' is invalid", r)
		}

		if len(r) >= 2 && r[0:2] == "- " {
			r = r[2:]
		}

		lastIndex := strings.LastIndex(r, "=")
		registryRegex := r[:lastIndex]
		project := r[lastIndex+2:]

		res = append(res, Rule{
			RegistryRegex: registryRegex,
			Project:       project,
			ServerURL:     u.Host,
		})
	}

	return res, nil
}

// MergeRules appends rule l after h, so h will be checked first.
// we append instead of merge since rules are regex, hard to merge,
// for example 'google.com' and '$google.com^' are the same.
func MergeRules(h, l []Rule) []Rule {
	return append(h, l...)
}
