package rule

import "strings"

type Rule struct {
	RegistryRegex string
	Project       string
	ServerURL     string
}

// assume rule.Rules are concatentated by ','.
func StringToRules(raw []string, server string) []Rule {
	res := make([]Rule, 0)

	for _, r := range raw {
		registryRegex := r[:strings.LastIndex(r, ",")]
		project := r[strings.LastIndex(r, ",")+1:]

		res = append(res, Rule{
			RegistryRegex: registryRegex,
			Project:       project,
			ServerURL:     server,
		})
	}

	return res
}

// append l after h, so l will be checked first.
// there could be cases that regex in h is `gcr.io`, while in l is `gcr.io*`.
func MergeRules(l, h []Rule) []Rule {
	return append(h, l...)
}
