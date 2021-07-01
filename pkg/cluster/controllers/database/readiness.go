package database

import (
	"context"
	"fmt"
	"os"
	"strings"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/database/api"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/pkg/errors"
	acidzalando "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do"
	acidzalandov1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	labels1 "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	CoreDatabase         = "core"
	NotaryServerDatabase = "notaryserver"
	NotarySignerDatabase = "notarysigner"
	DefaultDatabaseUser  = "harbor"
	PsqlRunningStatus    = "Running"
)

// Readiness reconcile will check postgre sql cluster if that has available.
// It does:
// - create postgre connection pool
// - ping postgre server
// - return postgre properties if postgre has available.
func (p *PostgreSQLController) Readiness(ctx context.Context, harborcluster *goharborv1.HarborCluster, curUnstructured *unstructured.Unstructured) (*lcm.CRStatus, error) {
	var (
		conn *Connect
		err  error
	)

	name := harborcluster.Name

	var pg api.Postgresql
	if err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(curUnstructured.UnstructuredContent(), &pg); err != nil {
		return nil, err
	}

	if pg.Status.PostgresClusterStatus != PsqlRunningStatus {
		return databaseNotReadyStatus(
			"Database is not ready",
			fmt.Sprintf("psql is %s", pg.Status.PostgresClusterStatus),
		), nil
	}

	conn, err = p.GetInClusterDatabaseInfo(ctx, harborcluster)
	if err != nil {
		return nil, err
	}

	if err := p.DeployComponentSecret(ctx, conn, getDatabasePasswordRefName(name), harborcluster); err != nil {
		return nil, err
	}

	if err := p.SetRefSecretOwner(ctx, "postgres", harborcluster); err != nil {
		return nil, err
	}

	if err := p.SetRefSecretOwner(ctx, "standby", harborcluster); err != nil {
		return nil, err
	}

	p.Log.Info("Database is ready.", "namespace", harborcluster.Namespace, "name", name)

	properties := &lcm.Properties{}
	addProperties(name, conn, properties)

	return databaseReadyStatus(
		"Database is ready",
		"Harbor component database secrets are already create",
		*properties,
	), nil
}

func addProperties(name string, conn *Connect, properties *lcm.Properties) {
	db := getHarborDatabaseSpec(name, conn)
	properties.Add(lcm.DatabasePropertyName, db)
}

func getHarborDatabaseSpec(name string, conn *Connect) *goharborv1.HarborDatabaseSpec {
	return &goharborv1.HarborDatabaseSpec{
		PostgresCredentials: harbormetav1.PostgresCredentials{
			Username:    DefaultDatabaseUser,
			PasswordRef: getDatabasePasswordRefName(name),
		},
		Hosts: []harbormetav1.PostgresHostSpec{
			{
				Host: conn.Host,
				Port: InClusterDatabasePortInt32,
			},
		},
		SSLMode: harbormetav1.PostgresSSLModeDisable,
	}
}

func getDatabasePasswordRefName(name string) string {
	return fmt.Sprintf("%s-%s-%s", name, "database", "password")
}

// SetRefSecretOwner set owner to pg operator created secrets.
func (p *PostgreSQLController) SetRefSecretOwner(ctx context.Context, dbName string, harborcluster *goharborv1.HarborCluster) error {
	ns := harborcluster.Namespace
	secretName := GenInClusterPasswordSecretName(dbName, p.resourceName(harborcluster.Namespace, harborcluster.Name))

	secret := &corev1.Secret{}
	if err := p.Client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: ns}, secret); err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}

		return err
	}

	if err := controllerutil.SetControllerReference(harborcluster, secret, p.Scheme); err != nil {
		return err
	}

	if err := p.Client.Update(ctx, secret); err != nil {
		return err
	}

	return nil
}

// DeployComponentSecret deploy harbor component database secret.
func (p *PostgreSQLController) DeployComponentSecret(ctx context.Context, conn *Connect, secretName string, harborcluster *goharborv1.HarborCluster) error {
	ns := harborcluster.Namespace
	secret := &corev1.Secret{}
	sc := p.GetDatabaseSecret(conn, ns, secretName)

	if err := controllerutil.SetControllerReference(harborcluster, sc, p.Scheme); err != nil {
		return err
	}

	err := p.Client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: ns}, secret)
	if kerr.IsNotFound(err) {
		p.Log.Info("Creating Harbor Component Secret", "namespace", ns, "name", secretName)

		if err := p.Client.Create(ctx, sc); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

// GetInClusterDatabaseInfo returns inCluster database connection client.
func (p *PostgreSQLController) GetInClusterDatabaseInfo(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*Connect, error) {
	var (
		connect *Connect
		err     error
	)

	pw, err := p.GetInClusterDatabasePassword(ctx, harborcluster)
	if err != nil {
		return connect, err
	}

	if connect, err = p.GetInClusterDatabaseConn(ctx, harborcluster, pw); err != nil {
		return connect, err
	}

	return connect, nil
}

// GetInClusterDatabaseConn returns inCluster database connection info.
func (p *PostgreSQLController) GetInClusterDatabaseConn(ctx context.Context, harborcluster *goharborv1.HarborCluster, pw string) (*Connect, error) {
	host, err := p.GetInClusterHost(ctx, harborcluster)
	if err != nil {
		return nil, err
	}

	conn := &Connect{
		Host:     host,
		Port:     InClusterDatabasePort,
		Password: pw,
		Username: DefaultDatabaseUser,
		Database: CoreDatabase,
	}

	return conn, nil
}

type SecretsTemplate string

func (f *SecretsTemplate) Format(a ...string) string {
	res := string(*f)

	for i := 0; i < len(a); i += 2 {
		res = strings.ReplaceAll(res, "{"+a[i]+"}", a[i+1])
	}

	return res
}

func GetPostgresSecretsTemplate() SecretsTemplate {
	template := SecretsTemplate(os.Getenv("POSTGRES_SECRETS_TEMPLATE"))

	if len(template) == 0 {
		return "{username}.{cluster}.credentials"
	}

	return template
}

func GenInClusterPasswordSecretName(user, crName string) string {
	template := GetPostgresSecretsTemplate()

	return template.Format(
		"username", strings.ReplaceAll(user, "_", "-"),
		"cluster", crName,
		"tprkind", acidzalandov1.PostgresCRDResourceKind,
		"tprgroup", acidzalando.GroupName)
}

// GetInClusterHost returns the Database master pod ip or service name.
func (p *PostgreSQLController) GetInClusterHost(ctx context.Context, harborcluster *goharborv1.HarborCluster) (string, error) {
	var (
		url string
		err error
	)

	_, err = rest.InClusterConfig()
	if err != nil {
		url, err = p.GetMasterPodsIP(ctx, harborcluster)
		if err != nil {
			return url, err
		}
	} else {
		url = fmt.Sprintf("%s.%s.svc", p.resourceName(harborcluster.Namespace, harborcluster.Name), harborcluster.Namespace)
	}

	return url, nil
}

// GetInClusterDatabasePassword is get inCluster postgresql password.
func (p *PostgreSQLController) GetInClusterDatabasePassword(ctx context.Context, harborcluster *goharborv1.HarborCluster) (string, error) {
	var pw string

	secretName := GenInClusterPasswordSecretName(DefaultDatabaseUser, p.resourceName(harborcluster.Namespace, harborcluster.Name))

	secret, err := p.GetSecret(ctx, harborcluster.Namespace, secretName)
	if err != nil {
		return pw, err
	}

	for k, v := range secret {
		if k == InClusterDatabasePasswordKey {
			pw = string(v)

			return pw, nil
		}
	}

	return pw, nil
}

// GetStatefulSetPods returns the postgresql master pod.
func (p *PostgreSQLController) GetStatefulSetPods(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*corev1.PodList, error) {
	resName := p.resourceName(harborcluster.Namespace, harborcluster.Name)

	label := map[string]string{
		"application":  "spilo",
		"cluster-name": resName,
		"spilo-role":   "master",
	}

	opts := &client.ListOptions{}
	set := labels1.SelectorFromSet(label)
	opts.LabelSelector = set
	pod := &corev1.PodList{}

	if err := p.Client.List(ctx, pod, opts); err != nil {
		p.Log.Error(err, "fail to get pod.",
			"namespace", harborcluster.Namespace, "name", resName)

		return nil, err
	}

	return pod, nil
}

// GetMasterPodsIP returns postgresql master node ip.
func (p *PostgreSQLController) GetMasterPodsIP(ctx context.Context, harborcluster *goharborv1.HarborCluster) (string, error) {
	var masterIP string

	podList, err := p.GetStatefulSetPods(ctx, harborcluster)
	if err != nil {
		return masterIP, err
	}

	if len(podList.Items) > 1 {
		return masterIP, errors.New("the number of master node copies cannot exceed 1")
	}

	for _, p := range podList.Items {
		if p.DeletionTimestamp != nil {
			continue
		}

		masterIP = p.Status.PodIP
	}

	return masterIP, nil
}
