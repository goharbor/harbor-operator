package controller

import (
	"context"
	"fmt"
	"strings"
)

func (c *Controller) NormalizeName(ctx context.Context, name string, suffixes ...string) string {
	name = fmt.Sprintf("%s-%s", name, c.GetName())

	if len(suffixes) > 0 {
		name += fmt.Sprintf("-%s", strings.Join(suffixes, "-"))
	}

	return name
}
