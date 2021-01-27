package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
)

func (c *Controller) NormalizeName(ctx context.Context, name string, suffixes ...string) string {
	name = fmt.Sprintf("%s-%s", name, c.GetName())

	if len(suffixes) > 0 {
		name += fmt.Sprintf("-%s", strings.Join(suffixes, "-"))
	}

	return name
}

func (c *Controller) GetClassName(ctx context.Context) (string, error) {
	className := ""

	configItem, err := configstore.Filter().Store(c.ConfigStore).Slice(config.HarborClassKey).GetFirstItem()
	if err != nil {
		if !config.IsNotFound(err, config.HarborClassKey) {
			return "", errors.Wrap(err, "cannot get config template path")
		}
	} else {
		className, err = configItem.Value()
		if err != nil {
			return "", errors.Wrap(err, "invalid config template path")
		}
	}

	return className, nil
}
