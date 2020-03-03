package notarysigner

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
	signerConfigKey = "signer.json"
)

var (
	once         sync.Once
	signerConfig []byte
)

func (r *Reconciler) InitConfigMaps() error {
	// We can't use a constant containing file path. Pkger don't understant if it's not the value passed as parameter.
	// const templatePath = "/my/Path"
	// pkger.Open(templatePath) --> Doesn't work
	signerFile, signerErr := pkger.Open("/assets/templates/notary/signer.json")
	if signerErr != nil {
		return errors.Wrapf(signerErr, "cannot open Notary Signer configuration template %s", "/assets/templates/notary/signer.json")
	}
	defer signerFile.Close()

	signerConfig, signerErr = ioutil.ReadAll(signerFile)
	if signerErr != nil {
		return errors.Wrapf(signerErr, "cannot read Notary Signer configuration template %s", "/assets/templates/notary/signer.json")
	}

	return nil
}

func (r *Reconciler) GetConfigMap(ctx context.Context, notary *goharborv1alpha2.NotarySigner) (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-notarysigner", notary.GetName()),
			Namespace: notary.GetNamespace(),
		},
		BinaryData: map[string][]byte{
			signerConfigKey: signerConfig,
		},
	}, nil
}
