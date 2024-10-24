package setup

import (
	"context"
	"fmt"

	"github.com/ovh/configstore"
	goharborv1 "github.com/plotly/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/plotly/harbor-operator/controllers"
	"github.com/plotly/harbor-operator/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	WebhookDisabledSuffixConfigKey = "webhook-disabled"
)

var webhooksBuilder = map[controllers.Controller][]WebHook{
	controllers.Core:               {&goharborv1.Core{}},
	controllers.Exporter:           {&goharborv1.Exporter{}},
	controllers.Harbor:             {&goharborv1.Harbor{}},
	controllers.JobService:         {&goharborv1.JobService{}},
	controllers.Registry:           {&goharborv1.Registry{}},
	controllers.Portal:             {&goharborv1.Portal{}},
	controllers.RegistryController: {&goharborv1.RegistryController{}},
	controllers.Trivy:              {&goharborv1.Trivy{}},
	controllers.HarborCluster:      {&goharborv1.HarborCluster{}},
	controllers.HarborProject:      {&goharborv1.HarborProject{}},
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
	configKey := fmt.Sprintf("%s-%s", wh.Name, WebhookDisabledSuffixConfigKey)

	ok, err := configstore.GetItemValueBool(configKey)
	if err == nil {
		return ok, nil
	}

	if config.IsNotFound(err, configKey) {
		return true, nil
	}

	return false, err
}
