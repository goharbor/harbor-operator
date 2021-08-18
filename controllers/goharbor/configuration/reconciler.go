package configuration

import (
	"context"
	"fmt"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	pkgharbor "github.com/goharbor/harbor-operator/pkg/harbor"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// New HarborConfiguration reconciler.
func New(ctx context.Context, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}
	r.Controller = commonCtrl.NewController(ctx, controllers.HarborCluster, nil, configStore)

	return r, nil
}

// Reconciler reconciles a configuration cr.
type Reconciler struct {
	*commonCtrl.Controller
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=goharbor.io,resources=harborconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=goharbor.io,resources=harborconfigurations/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&goharborv1.HarborConfiguration{}).
		Complete(r)
}

func (r *Reconciler) NormalizeName(ctx context.Context, name string, suffixes ...string) string {
	suffixes = append([]string{"Configuration"}, suffixes...)

	return strings.NormalizeName(name, suffixes...)
}

// Reconcile does configuration reconcile.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) { // nolint:funlen
	log := r.Log.WithValues("resource", req.NamespacedName)

	log.Info("Start reconciling")

	hc := &goharborv1.HarborConfiguration{}
	if err = r.Client.Get(ctx, req.NamespacedName, hc); err != nil {
		if apierrors.IsNotFound(err) {
			// The resource may have be deleted after reconcile request coming in
			// Reconcile is done
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("error get harbor configuration: %w", err)
	}

	hcCopy := hc.DeepCopy()

	defer func() {
		if err != nil {
			hc.Status.Status = goharborv1.HarborConfigurationStatusFail
		} else {
			hc.Status.Status = goharborv1.HarborConfigurationStatusReady
			now := metav1.Now()
			hc.Status.LastApplyTime = &now
			hc.Status.LastConfiguration = &hcCopy.Spec
		}

		log.Info("Reconcile end", "result", res, "error", err, "updateStatusError", r.Client.Status().Update(ctx, hc))
	}()

	hc.Status.Status = goharborv1.HarborConfigurationStatusUnknown

	// get harbor cr
	harborCluster := &goharborv1.HarborCluster{}
	if err = r.Client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: hc.Spec.HarborClusterRef}, harborCluster); err != nil {
		err = fmt.Errorf("error get harborCluster: %w", err)
		hc.Status.Reason = "HarborClusterError"

		return
	}
	// get harbor client
	harborClient, err := r.getHarborClient(ctx, harborCluster)
	if err != nil {
		err = fmt.Errorf("error get harbor client: %w", err)
		hc.Status.Reason = "HarborClientError"

		return
	}
	// assemble hc
	payload, err := r.assembleHarborConfiguration(ctx, hc)
	if err != nil {
		err = fmt.Errorf("error assemble harbor configuration: %w", err)
		hc.Status.Reason = "AssembleConfigurationError"

		return
	}
	// apply configuration
	if err = harborClient.ApplyConfiguration(ctx, payload); err != nil {
		err = fmt.Errorf("apply harbor configuration error: %w", err)
		hc.Status.Reason = "ApplyConfigurationError"

		return
	}

	return ctrl.Result{}, nil
}

// getHarborClient gets harbor client.
func (r *Reconciler) getHarborClient(ctx context.Context, harbor *goharborv1.HarborCluster) (pkgharbor.Client, error) {
	url := harbor.Spec.ExternalURL
	if len(url) == 0 {
		return nil, errors.Errorf("harbor url is invalid")
	}

	var opts []pkgharbor.ClientOption

	adminSecretRef := harbor.Spec.HarborAdminPasswordRef
	if len(adminSecretRef) > 0 {
		// fetch admin password
		secret := &corev1.Secret{}
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: harbor.Namespace, Name: adminSecretRef}, secret); err != nil {
			return nil, fmt.Errorf("error get harbor admin secret: %w", err)
		}

		password := string(secret.Data["secret"])
		opts = append(opts, pkgharbor.WithCredential("admin", password))
	}

	return pkgharbor.NewClient(url, opts...), nil
}

// assembleConfig assembles password filed from secret.
func (r *Reconciler) assembleHarborConfiguration(ctx context.Context, hc *goharborv1.HarborConfiguration) (payload []byte, err error) { // nolint:funlen
	secretValueGetter := func(secretName, secretNamespace, key string) (string, error) {
		secret := &corev1.Secret{}
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: secretNamespace, Name: secretName}, secret); err != nil {
			return "", err
		}

		if v, ok := secret.Data[key]; ok {
			return string(v), nil
		}

		return "", errors.Errorf("secret key '%s' not found in secret data", key)
	}

	// "email_password", "ldap_search_password", "uaa_client_secret", "oidc_client_secret"
	// these configuration spec need extracts value from secret.

	if len(hc.Spec.Configuration.EmailPassword) != 0 {
		password, err := secretValueGetter(hc.Spec.Configuration.EmailPassword, hc.Namespace, "email_password")
		if err != nil {
			return nil, fmt.Errorf("error extract email_password from secret %s: %w", hc.Spec.Configuration.EmailPassword, err)
		}

		hc.Spec.Configuration.EmailPassword = password
	}

	if len(hc.Spec.Configuration.LdapSearchPassword) != 0 {
		password, err := secretValueGetter(hc.Spec.Configuration.LdapSearchPassword, hc.Namespace, "ldap_search_password")
		if err != nil {
			return nil, fmt.Errorf("error extract ldap_search_password from secret %s: %w", hc.Spec.Configuration.LdapSearchPassword, err)
		}

		hc.Spec.Configuration.LdapSearchPassword = password
	}

	if len(hc.Spec.Configuration.UaaClientSecret) != 0 {
		secret, err := secretValueGetter(hc.Spec.Configuration.UaaClientSecret, hc.Namespace, "uaa_client_secret")
		if err != nil {
			return nil, fmt.Errorf("error extract uaa_client_secret from secret %s: %w", hc.Spec.Configuration.UaaClientSecret, err)
		}

		hc.Spec.Configuration.UaaClientSecret = secret
	}

	if len(hc.Spec.Configuration.OidcClientSecret) != 0 {
		secret, err := secretValueGetter(hc.Spec.Configuration.OidcClientSecret, hc.Namespace, "oidc_client_secret")
		if err != nil {
			return nil, fmt.Errorf("error extract oidc_client_secret from secret %s: %w", hc.Spec.Configuration.UaaClientSecret, err)
		}

		hc.Spec.Configuration.OidcClientSecret = secret
	}
	// convert spec config to json format
	p, err := hc.Spec.Configuration.ToJSON()
	if err != nil {
		return nil, err
	}
	// from json payload to harbor configuration
	c, err := pkgharbor.FromJSONToConfiguration(p)
	if err != nil {
		return nil, err
	}

	return c.Payload()
}
