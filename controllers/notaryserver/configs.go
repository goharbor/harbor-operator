package notaryserver

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	serverConfigKey = "server.json"
)

var (
	once          sync.Once
	configContent []byte
)

func (r *Reconciler) InitConfigMaps() error {
	// We can't use a constant containing file path. Pkger don't understant if it's not the value passed as parameter.
	// const templatePath = "/my/Path"
	// pkger.Open(templatePath) --> Doesn't work
	serverFile, serverErr := pkger.Open("/assets/templates/notary/server.json")
	if serverErr != nil {
		return errors.Wrapf(serverErr, "cannot open Notary Server configuration template %s", "/assets/templates/notary/server.json")
	}
	defer serverFile.Close()

	configContent, serverErr = ioutil.ReadAll(serverFile)
	if serverErr != nil {
		return errors.Wrapf(serverErr, "cannot read Notary Server configuration template %s", "/assets/templates/notary/server.json")
	}

	return nil
}

func (r *Reconciler) GetConfigMap(ctx context.Context, notary *goharborv1alpha2.NotaryServer) (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-notaryserver", notary.GetName()),
			Namespace: notary.GetNamespace(),
		},
		BinaryData: map[string][]byte{
			serverConfigKey: configContent,
		},
	}, nil
}
