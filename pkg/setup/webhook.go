package setup

import (
	"context"
	"fmt"

	"github.com/ovh/configstore"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	WebhookDisabledSuffixConfigKey = "webhook-disabled"
)

type WebHook interface {
	SetupWebhookWithManager(context.Context, manager.Manager) error
}

type webHook struct {
	Name    ControllerUID
	webhook WebHook
}

func (wh *webHook) WithManager(ctx context.Context, mgr manager.Manager) error {
	if wh.webhook != nil {
		return wh.webhook.SetupWebhookWithManager(ctx, mgr)
	}

	return nil
}

func (wh *webHook) IsEnabled(ctx context.Context) (bool, error) {
	ok, err := configstore.GetItemValueBool(fmt.Sprintf("%s-%s", wh.Name, WebhookDisabledSuffixConfigKey))
	if err == nil {
		return ok, nil
	}

	if _, ok := err.(configstore.ErrItemNotFound); ok {
		return true, nil
	}

	return false, err
}
