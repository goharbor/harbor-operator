package setup

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/webhooks/harborserverconfiguration"
	"github.com/goharbor/harbor-operator/webhooks/pod"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	kauthn "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func WithManager(ctx context.Context, mgr manager.Manager) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return errors.Wrap(ControllersWithManager(ctx, mgr), "controllers")
	})

	g.Go(func() error {
		return errors.Wrap(WebhooksWithManager(ctx, mgr), "webhooks")
	})

	return g.Wait()
}

func populateContext(ctx context.Context, mgr manager.Manager) (context.Context, error) {
	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(mgr.GetConfig())

	preferredResources, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return ctx, err
	}

	resources := discovery.FilteredBy(discovery.ResourcePredicateFunc(func(groupVersion string, r *metav1.APIResource) bool {
		check := sets.NewString([]string(r.Verbs)...).HasAll("delete", "list", "create")
		check = check && r.Namespaced

		return check
	}), preferredResources)
	deletableResources := make(map[schema.GroupVersionKind]struct{})

	for _, rl := range resources {
		gv, err := schema.ParseGroupVersion(rl.GroupVersion)
		if err != nil {
			return ctx, err
		}

		for i := range rl.APIResources {
			sar := &kauthn.SelfSubjectAccessReview{
				Spec: kauthn.SelfSubjectAccessReviewSpec{
					ResourceAttributes: &kauthn.ResourceAttributes{
						Verb:     "delete",
						Group:    gv.Group,
						Version:  gv.Version,
						Resource: rl.APIResources[i].Name,
					},
				},
			}
			if err := mgr.GetClient().Create(ctx, sar); err != nil {
				return ctx, err
			}

			if sar.Status.Allowed {
				deletableResources[gv.WithKind(rl.APIResources[i].Kind)] = struct{}{}
			}
		}
	}

	application.SetDeletableResources(&ctx, deletableResources)

	return ctx, nil
}

func ControllersWithManager(ctx context.Context, mgr manager.Manager) error {
	ctx, err := populateContext(ctx, mgr)
	if err != nil {
		return errors.Wrap(err, "populateContext")
	}

	var g errgroup.Group

	for name, builder := range controllersBuilder {
		ctx := ctx

		logger.Set(&ctx, logger.Get(ctx).WithName(name.String()))

		c := NewController(name, builder)

		ok, err := c.IsEnabled(ctx)
		if err != nil {
			return errors.Wrap(err, "cannot check if controller is enabled")
		}

		if !ok {
			logger.Get(ctx).Info("Controller disabled")

			continue
		}

		name := name

		g.Go(func() error {
			_, err := c.WithManager(ctx, mgr)

			return errors.Wrap(err, name.String())
		})
	}

	return g.Wait()
}

func WebhooksWithManager(ctx context.Context, mgr manager.Manager) error {
	for name, webhooks := range webhooksBuilder {
		ctx := ctx

		for _, webhook := range webhooks {
			logger.Set(&ctx, logger.Get(ctx).WithName(name.String()))

			wh := &webHook{
				Name:    name,
				webhook: webhook,
			}

			ok, err := wh.IsEnabled(ctx)
			if err != nil {
				return errors.Wrap(err, "cannot check if webhook is enabled")
			}

			if !ok {
				logger.Get(ctx).Info("Webhook disabled")

				continue
			}

			// Fail earlier.
			// 'controller-runtime' does not support webhook registrations concurrently.
			// Check issue: https://github.com/kubernetes-sigs/controller-runtime/issues/1285.
			if err := wh.WithManager(ctx, mgr); err != nil {
				return errors.Wrap(err, name.String())
			}
		}
	}

	// setup separate webhooks
	setupCustomWebhooks(mgr)

	return nil
}

func setupCustomWebhooks(mgr manager.Manager) {
	mgr.GetWebhookServer().Register("/mutate-image-path", &webhook.Admission{
		Handler: &pod.ImagePathRewriter{
			Client: mgr.GetClient(),
			Log:    logf.Log.WithName("webhooks").WithName("MutatingImagePath"),
		},
	})

	mgr.GetWebhookServer().Register("/validate-hsc", &webhook.Admission{
		Handler: &harborserverconfiguration.Validator{
			Client: mgr.GetClient(),
			Log:    logf.Log.WithName("webhooks").WithName("HarborServerConfigurationValidator"),
		},
	})
}
