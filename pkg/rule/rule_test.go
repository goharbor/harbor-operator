package rule_test

import (
	"testing"

	"github.com/goharbor/harbor-operator/pkg/rule"
	"github.com/stretchr/testify/require"
)

var (
	testURL = "ww.test.com"
	rawURL  = "https://ww.test.com"
)

func Test_StringToRules(t *testing.T) {
	type testcase struct {
		description   string
		rules         []string
		expectedRules []rule.Rule
	}

	tests := []testcase{
		{
			description: "rules from hsc",
			rules:       []string{"docker.io=>harborproject1", "*=>harborproject2", "quay.io=>harborproject3"},
			expectedRules: []rule.Rule{
				{
					RegistryRegex: "docker.io",
					Project:       "harborproject1",
					ServerURL:     testURL,
				},
				{
					RegistryRegex: "*",
					Project:       "harborproject2",
					ServerURL:     testURL,
				},
				{
					RegistryRegex: "quay.io",
					Project:       "harborproject3",
					ServerURL:     testURL,
				},
			},
		},
		{
			description: "rules from configMap",
			rules:       []string{"- docker.io=>harborproject1", "- *=>harborproject2", "- quay.io=>harborproject3"},
			expectedRules: []rule.Rule{
				{
					RegistryRegex: "docker.io",
					Project:       "harborproject1",
					ServerURL:     testURL,
				},
				{
					RegistryRegex: "*",
					Project:       "harborproject2",
					ServerURL:     testURL,
				},
				{
					RegistryRegex: "quay.io",
					Project:       "harborproject3",
					ServerURL:     testURL,
				},
			},
		},
	}

	for _, tc := range tests {
		output, err := rule.StringToRules(tc.rules, rawURL)
		require.Nil(t, err)
		require.Equal(t, tc.expectedRules, output, tc.description)
	}
}

func Test_MergeRules(t *testing.T) {
	type testcase struct {
		description   string
		rules1        []rule.Rule
		rules2        []rule.Rule
		expectedRules []rule.Rule
	}

	tests := []testcase{
		{
			description: "rules from hsc",
			rules1: []rule.Rule{
				{
					RegistryRegex: "docker.io",
					Project:       "harborproject1",
					ServerURL:     testURL,
				},
			},
			rules2: []rule.Rule{
				{
					RegistryRegex: "quay.io",
					Project:       "harborproject3",
					ServerURL:     testURL,
				},
			},
			expectedRules: []rule.Rule{
				{
					RegistryRegex: "docker.io",
					Project:       "harborproject1",
					ServerURL:     testURL,
				},
				{
					RegistryRegex: "quay.io",
					Project:       "harborproject3",
					ServerURL:     testURL,
				},
			},
		},
	}

	for _, tc := range tests {
		output := rule.MergeRules(tc.rules1, tc.rules2)

		require.Equal(t, tc.expectedRules, output, tc.description)
	}
}
