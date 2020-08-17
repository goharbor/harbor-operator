package setup

import (
	"context"
	"fmt"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/ovh/configstore"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	WebhookDisabledSuffixConfigKey = "webhook-disabled"
)

var webhooksBuilder = map[controllers.Controller]WebHook{
	controllers.Harbor:       &goharborv1alpha2.Harbor{},
	controllers.JobService:   &goharborv1alpha2.JobService{},
	controllers.Registry:     &goharborv1alpha2.Registry{},
	controllers.NotaryServer: &goharborv1alpha2.NotaryServer{},
	controllers.NotarySigner: &goharborv1alpha2.NotarySigner{},
}

type WebHook interface {
	SetupWebhookWithManager(context.Context, manager.Manager) error
}

type webHook struct {
	Name    controllers.Controller
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
