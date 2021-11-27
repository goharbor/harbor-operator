package v1beta1

import (
	"context"

	traitWebhook "github.com/goharbor/harbor-operator/controllers/goharbor/trait/webhook"
	"github.com/spotahome/redis-operator/log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (trait *HarborClusterTrait) SetupWebhookWithManager(_ context.Context, mgr ctrl.Manager) error {
	log.Info("start to launch harbor-cluster-trait webhook")

	hookServer := mgr.GetWebhookServer()
	hookServer.Register("/harbor-cluster-trait-mutate-v1-pod", &webhook.Admission{Handler: &traitWebhook.PodAnnotator{Client: mgr.GetClient()}})

	log.Info("success to launch harbor-cluster-trait webhook")

	return nil
}
