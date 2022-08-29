package configuration

import (
	"context"
	"encoding/json"

	"github.com/goharbor/go-client/pkg/harbor"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/configure"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
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
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) { //nolint:funlen
	log := r.Log.WithValues("resource", req.NamespacedName)

	log.Info("Start reconciling")

	hc := &goharborv1.HarborConfiguration{}
	if err = r.Client.Get(ctx, req.NamespacedName, hc); err != nil {
		if apierrors.IsNotFound(err) {
			// The resource may have be deleted after reconcile request coming in
			// Reconcile is done
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, errors.Wrapf(err, "error get harbor configuration %v", req)
	}

	hcCopy := hc.DeepCopy()

	defer func() {
		if err != nil {
			hc.Status.Status = goharborv1.HarborConfigurationStatusFail
			hc.Status.Message = err.Error()
		} else {
			hc.Status.Status = goharborv1.HarborConfigurationStatusReady
			hc.Status.Reason = ""
			hc.Status.Message = ""
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
		err = errors.Wrapf(err, "error get harborCluster %s", hc.Spec.HarborClusterRef)
		hc.Status.Reason = "HarborClusterError"

		return
	}
	// get harbor client
	harborClient, err := r.getHarborClient(ctx, harborCluster)
	if err != nil {
		err = errors.Wrapf(err, "error get harbor client")
		hc.Status.Reason = "HarborClientError"

		return
	}
	// assemble hc
	configurationModel, err := r.assembleHarborConfiguration(ctx, hc)
	if err != nil {
		err = errors.Wrapf(err, "error assemble harbor configuration")
		hc.Status.Reason = "AssembleConfigurationError"

		return
	}
	// apply configuration
	params := configure.NewUpdateConfigurationsParams().WithConfigurations(configurationModel)
	if _, err = harborClient.V2().Configure.UpdateConfigurations(ctx, params); err != nil {
		err = errors.Wrapf(err, "error apply harbor configuration")
		hc.Status.Reason = "ApplyConfigurationError"

		return
	}

	return ctrl.Result{}, nil
}

// getHarborClient gets harbor client.
func (r *Reconciler) getHarborClient(ctx context.Context, hc *goharborv1.HarborCluster) (*harbor.ClientSet, error) {
	var (
		username = "admin"
		password = ""
	)

	adminSecretRef := hc.Spec.HarborAdminPasswordRef
	if len(adminSecretRef) > 0 {
		// fetch admin password
		secret := &corev1.Secret{}
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: hc.Namespace, Name: adminSecretRef}, secret); err != nil {
			return nil, errors.Wrapf(err, "failed to get harbor admin secret: %s", adminSecretRef)
		}

		password = string(secret.Data["secret"])
	}

	config := harbor.ClientSetConfig{
		URL:      hc.Spec.ExternalURL,
		Username: username,
		Password: password,
	}

	return harbor.NewClientSet(&config)
}

// assembleConfig assembles password filed from secret.
func (r *Reconciler) assembleHarborConfiguration(ctx context.Context, hc *goharborv1.HarborConfiguration) (model *models.Configurations, err error) { //nolint:funlen
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
			return nil, errors.Wrapf(err, "error extract email_password from secret %s", hc.Spec.Configuration.EmailPassword)
		}

		hc.Spec.Configuration.EmailPassword = password
	}

	if len(hc.Spec.Configuration.LdapSearchPassword) != 0 {
		password, err := secretValueGetter(hc.Spec.Configuration.LdapSearchPassword, hc.Namespace, "ldap_search_password")
		if err != nil {
			return nil, errors.Wrapf(err, "error extract ldap_search_password from secret %s", hc.Spec.Configuration.LdapSearchPassword)
		}

		hc.Spec.Configuration.LdapSearchPassword = password
	}

	if len(hc.Spec.Configuration.UaaClientSecret) != 0 {
		secret, err := secretValueGetter(hc.Spec.Configuration.UaaClientSecret, hc.Namespace, "uaa_client_secret")
		if err != nil {
			return nil, errors.Wrapf(err, "error extract uaa_client_secret from secret %s", hc.Spec.Configuration.UaaClientSecret)
		}

		hc.Spec.Configuration.UaaClientSecret = secret
	}

	if len(hc.Spec.Configuration.OidcClientSecret) != 0 {
		secret, err := secretValueGetter(hc.Spec.Configuration.OidcClientSecret, hc.Namespace, "oidc_client_secret")
		if err != nil {
			return nil, errors.Wrapf(err, "error extract oidc_client_secret from secret %s", hc.Spec.Configuration.UaaClientSecret)
		}

		hc.Spec.Configuration.OidcClientSecret = secret
	}
	// convert spec config to json format
	p, err := hc.Spec.Configuration.ToJSON()
	if err != nil {
		return nil, err
	}

	model = &models.Configurations{}
	if err = json.Unmarshal(p, model); err != nil {
		return nil, err
	}

	return model, nil
}
