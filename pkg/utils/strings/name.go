package strings

import (
	"fmt"
	"strings"
)

const NormalizationSeparator = "-"

func NormalizeName(name string, suffixes ...string) string {
	if len(suffixes) > 0 {
		name += fmt.Sprintf("%s%s", NormalizationSeparator, strings.Join(suffixes, NormalizationSeparator))
	}

	return name
}
