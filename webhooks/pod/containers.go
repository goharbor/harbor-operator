package pod

import (
	"fmt"
	"strings"

	"github.com/containers/image/v5/docker/reference"
	"github.com/goharbor/harbor-operator/pkg/rule"
	"github.com/umisama/go-regexpcache"
)

const BareRegistry = "docker.io"

// RegistryFromImageRef returns the registry (and port, if set) from the image reference,
// otherwise returns the default bare registry, "docker.io".
func RegistryFromImageRef(imageReference string) (registry string, err error) {
	ref, err := reference.ParseDockerRef(imageReference)
	if err != nil {
		return "", err
	}

	return reference.Domain(ref), nil
}

// ReplaceRegistryInImageRef returns the the image reference with the registry replaced.
func ReplaceRegistryInImageRef(imageReference, replacementRegistry string) (imageRef string, err error) {
	named, err := reference.ParseDockerRef(imageReference)
	if err != nil {
		return "", err
	}

	return strings.Replace(named.String(), reference.Domain(named), replacementRegistry, 1), nil
}

// rewriteContainer replaces any registries matching the image rules with the given serverURL.
func rewriteContainer(imageReference string, rules []rule.Rule) (imageRef string, err error) {
	registry, err := RegistryFromImageRef(imageReference)
	if err != nil {
		return "", err
	}

	var starRule *rule.Rule

	for i, r := range rules {
		if r.RegistryRegex != "*" {
			regex, err := regexpcache.Compile(r.RegistryRegex)
			if err != nil {
				return "", err
			}

			if regex.MatchString(registry) {
				rewritten := fmt.Sprintf("%s/%s", r.ServerURL, r.Project)

				return ReplaceRegistryInImageRef(imageReference, rewritten)
			}
		} else {
			starRule = &rules[i]
		}
	}

	// * has the lowerest priority in the rules, match this in the end.
	if starRule != nil {
		rewritten := fmt.Sprintf("%s/%s", starRule.ServerURL, starRule.Project)

		return ReplaceRegistryInImageRef(imageReference, rewritten)
	}

	return "", nil
}
