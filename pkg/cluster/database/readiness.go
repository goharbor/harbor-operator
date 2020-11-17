package database

import (
	"errors"
	"fmt"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"

	"github.com/goharbor/harbor-operator/pkg/lcm"
	"github.com/jackc/pgx/v4"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	labels1 "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	HarborCore         = "core"
	HarborClair        = "clair"
	HarborNotaryServer = "notaryServer"
	HarborNotarySigner = "notarySigner"

	CoreDatabase         = "core"
	ClairDatabase        = "clair"
	NotaryServerDatabase = "notaryserver"
	NotarySignerDatabase = "notarysigner"

	CoreSecretName         = "core"
	ClairSecretName        = "clair"
	NotaryServerSecretName = "notary-server"
	NotarySignerSecretName = "notary-signer"
)

// Readiness reconcile will check postgre sql cluster if that has available.
// It does:
// - create postgre connection pool
// - ping postgre server
// - return postgre properties if postgre has available
func (p *PostgreSQLController) Readiness() (*lcm.CRStatus, error) {
	var (
		conn   *Connect
		client *pgx.Conn
		err    error
	)

	name := p.HarborCluster.Name

	conn, client, err = p.GetInClusterDatabaseInfo()
	if err != nil {
		return nil, err
	}

	defer client.Close(p.Ctx)

	if err := client.Ping(p.Ctx); err != nil {
		p.Log.Error(err, "Fail to check Database.",
			"namespace", p.HarborCluster.Namespace,
			"name", p.HarborCluster.Name)
		return nil, err
	}
	p.Log.Info("Database already ready.",
		"namespace", p.HarborCluster.Namespace,
		"name", p.HarborCluster.Name)

	properties := &lcm.Properties{}
	if err := p.DeployComponentSecret(conn, getDatabasePasswordRefName(name)); err != nil {
		return nil, err
	}
	addProperties(name, conn, properties)

	crStatus := lcm.New(goharborv1alpha2.DatabaseReady).
		WithStatus(corev1.ConditionTrue).
		WithReason("database already ready").
		WithMessage("harbor component database secrets are already create.").
		WithProperties(*properties)
	return crStatus, nil
}

func addProperties(name string, conn *Connect, properties *lcm.Properties) {
	db := getHarborDatabaseSpec(name, conn)
	properties.Add("database", db)
}

func getHarborDatabaseSpec(name string, conn *Connect) *goharborv1alpha2.HarborDatabaseSpec {
	return &goharborv1alpha2.HarborDatabaseSpec{
		PostgresCredentials: harbormetav1.PostgresCredentials{
			Username:    "zalando",
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

func getPropertyName(key string) string {
	return fmt.Sprintf("%sSecret", key)
}

func getComponentSecretName(component string) string {
	return fmt.Sprintf("%s-database", component)
}

// DeployComponentSecret deploy harbor component database secret
func (p *PostgreSQLController) DeployComponentSecret(conn *Connect, secretName string) error {

	secret := &corev1.Secret{}
	sc := p.GetDatabaseSecret(conn, secretName)

	if err := controllerutil.SetControllerReference(p.HarborCluster, sc, p.Scheme); err != nil {
		return err
	}
	err := p.Client.Get(types.NamespacedName{Name: secretName, Namespace: p.HarborCluster.Namespace}, secret)
	if kerr.IsNotFound(err) {
		p.Log.Info("Creating Harbor Component Secret",
			"namespace", p.HarborCluster.Namespace,
			"name", secretName)
		return p.Client.Create(sc)
	} else if err != nil {
		return err
	}
	return nil
}

// GetInClusterDatabaseInfo returns inCluster database connection client
func (p *PostgreSQLController) GetInClusterDatabaseInfo() (*Connect, *pgx.Conn, error) {
	var (
		connect *Connect
		client  *pgx.Conn
		err     error
	)

	pw, err := p.GetInClusterDatabasePassword()
	if err != nil {
		return connect, client, err
	}

	if connect, err = p.GetInClusterDatabaseConn(p.GetDatabaseName(), pw); err != nil {
		return connect, client, err
	}

	url := connect.GenDatabaseUrl()

	client, err = pgx.Connect(p.Ctx, url)
	if err != nil {
		p.Log.Error(err, "Unable to connect to database")
		return connect, client, err
	}

	return connect, client, nil
}

// GetInClusterDatabaseConn returns inCluster database connection info
func (p *PostgreSQLController) GetInClusterDatabaseConn(name, pw string) (*Connect, error) {
	host, err := p.GetInClusterHost(name)
	if err != nil {
		return nil, err
	}
	conn := &Connect{
		Host:     host,
		Port:     InClusterDatabasePort,
		Password: pw,
		Username: InClusterDatabaseUserName,
		Database: InClusterDatabaseName,
	}
	return conn, nil
}

func GenInClusterPasswordSecretName(teamID, name string) string {
	return fmt.Sprintf("postgres.%s-%s.credentials", teamID, name)
}

// GetInClusterHost returns the Database master pod ip or service name
func (p *PostgreSQLController) GetInClusterHost(name string) (string, error) {
	var (
		url string
		err error
	)
	_, err = rest.InClusterConfig()
	if err != nil {
		url, err = p.GetMasterPodsIP()
		if err != nil {
			return url, err
		}
	} else {
		url = fmt.Sprintf("%s.%s.svc", name, p.HarborCluster.Namespace)
	}

	return url, nil
}

func (p *PostgreSQLController) GetDatabaseName() string {
	return fmt.Sprintf("%s-%s", p.HarborCluster.Namespace, p.HarborCluster.Name)
}

// GetInClusterDatabasePassword is get inCluster postgresql password
func (p *PostgreSQLController) GetInClusterDatabasePassword() (string, error) {
	var pw string

	secretName := GenInClusterPasswordSecretName(p.HarborCluster.Namespace, p.HarborCluster.Name)
	secret, err := p.GetSecret(secretName)
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

// GetStatefulSetPods returns the postgresql master pod
func (p *PostgreSQLController) GetStatefulSetPods() (*corev1.PodList, error) {
	name := p.GetDatabaseName()
	label := map[string]string{
		"application":  "spilo",
		"cluster-name": name,
		"spilo-role":   "master",
	}

	opts := &client.ListOptions{}
	set := labels1.SelectorFromSet(label)
	opts.LabelSelector = set
	pod := &corev1.PodList{}

	if err := p.Client.List(opts, pod); err != nil {
		p.Log.Error(err, "fail to get pod.",
			"namespace", p.HarborCluster.Namespace, "name", name)
		return nil, err
	}
	return pod, nil
}

// GetMasterPodsIP returns postgresql master node ip
func (p *PostgreSQLController) GetMasterPodsIP() (string, error) {
	var masterIP string
	podList, err := p.GetStatefulSetPods()
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
