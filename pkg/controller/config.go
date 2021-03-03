package controller

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/config"
)

func (c *Controller) StringConfig(ctx context.Context, key string, defaultValue string) (string, error) {
	return config.GetString(c.ConfigStore, key, defaultValue)
}

func (c *Controller) IntConfig(ctx context.Context, key string, defaultValue int) (int, error) {
	return config.GetInt(c.ConfigStore, key, defaultValue)
}

func (c *Controller) BoolConfig(ctx context.Context, key string, defaultValue bool) (bool, error) {
	return config.GetBool(c.ConfigStore, key, defaultValue)
}
