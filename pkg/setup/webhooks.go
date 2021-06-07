package setup

import (
	"context"
	"fmt"

	goharborv1alpha3 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/ovh/configstore"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	WebhookDisabledSuffixConfigKey = "webhook-disabled"
)

var webhooksBuilder = map[controllers.Controller][]WebHook{
	controllers.ChartMuseum:        {&goharborv1alpha3.ChartMuseum{}},
	controllers.Core:               {&goharborv1alpha3.Core{}},
	controllers.Exporter:           {&goharborv1alpha3.Exporter{}},
	controllers.Harbor:             {&goharborv1alpha3.Harbor{}},
	controllers.JobService:         {&goharborv1alpha3.JobService{}},
	controllers.Registry:           {&goharborv1alpha3.Registry{}},
	controllers.Portal:             {&goharborv1alpha3.Portal{}},
	controllers.RegistryController: {&goharborv1alpha3.RegistryController{}},
	controllers.Trivy:              {&goharborv1alpha3.Trivy{}},
	controllers.NotaryServer:       {&goharborv1alpha3.NotaryServer{}},
	controllers.NotarySigner:       {&goharborv1alpha3.NotarySigner{}},
	controllers.HarborCluster:      {&goharborv1alpha3.HarborCluster{}},
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
