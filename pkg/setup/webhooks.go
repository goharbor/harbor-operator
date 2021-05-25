package setup

import (
	"context"
	"fmt"

	goharborv1alpha3 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	goharborv1beta1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/ovh/configstore"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	WebhookDisabledSuffixConfigKey = "webhook-disabled"
)

var webhooksBuilder = map[controllers.Controller][]WebHook{
	controllers.ChartMuseum:        {&goharborv1beta1.ChartMuseum{}},
	controllers.Core:               {&goharborv1beta1.Core{}},
	controllers.Exporter:           {&goharborv1beta1.Core{}},
	controllers.Harbor:             {&goharborv1alpha3.Harbor{}, &goharborv1beta1.Harbor{}},
	controllers.JobService:         {&goharborv1alpha3.JobService{}, &goharborv1beta1.JobService{}},
	controllers.Registry:           {&goharborv1alpha3.Registry{}, &goharborv1beta1.Registry{}},
	controllers.Portal:             {&goharborv1beta1.Portal{}},
	controllers.RegistryController: {&goharborv1beta1.RegistryController{}},
	controllers.Trivy:              {&goharborv1beta1.Trivy{}},
	controllers.NotaryServer:       {&goharborv1alpha3.NotaryServer{}},
	controllers.NotarySigner:       {&goharborv1alpha3.NotarySigner{}},
	controllers.HarborCluster:      {&goharborv1beta1.HarborCluster{}, &goharborv1alpha3.HarborCluster{}},
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
