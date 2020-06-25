package controller

import (
	"context"
	"fmt"
)

func (c *Controller) NormalizeName(ctx context.Context, name string) string {
	return fmt.Sprintf("%s-%s", name, c.GetName())
}
