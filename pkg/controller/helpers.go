package controller

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
)

func (c *Controller) NormalizeName(ctx context.Context, name string, suffixes ...string) string {
	suffixes = append([]string{c.GetName()}, suffixes...)

	return strings.NormalizeName(name, suffixes...)
}

const DefaultClassName = ""

func (c *Controller) GetClassName(ctx context.Context) (string, error) {
	return c.StringConfig(ctx, config.HarborClassKey, DefaultClassName)
}
