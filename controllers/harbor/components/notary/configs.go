package notary

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ovh/harbor-operator/pkg/factories/application"
)

const (
	serverConfigKey = "server.json"
	signerConfigKey = "signer.json"
)

var (
	once         sync.Once
	serverConfig []byte
	signerConfig []byte
)

func InitConfigMaps() {
	// We can't use a constant containing file path. Pkger don't understant if it's not the value passed as parameter.
	// const templatePath = "/my/Path"
	// pkger.Open(templatePath) --> Doesn't work
	serverFile, serverErr := pkger.Open("/assets/templates/notary/server.json")
	if serverErr != nil {
		panic(errors.Wrapf(serverErr, "cannot open Notary Server configuration template %s", "/assets/templates/notary/server.json"))
	}
	defer serverFile.Close()

	serverConfig, serverErr = ioutil.ReadAll(serverFile)
	if serverErr != nil {
		panic(errors.Wrapf(serverErr, "cannot read Notary Server configuration template %s", "/assets/templates/notary/server.json"))
	}

	signerFile, signerErr := pkger.Open("/assets/templates/notary/signer.json")
	if signerErr != nil {
		panic(errors.Wrapf(signerErr, "cannot open Notary Signer configuration template %s", "/assets/templates/notary/signer.json"))
	}
	defer signerFile.Close()

	signerConfig, signerErr = ioutil.ReadAll(signerFile)
	if signerErr != nil {
		panic(errors.Wrapf(signerErr, "cannot read Notary Signer configuration template %s", "/assets/templates/notary/signer.json"))
	}
}

func (n *Notary) GetConfigMaps(ctx context.Context) []*corev1.ConfigMap {
	once.Do(InitConfigMaps)

	operatorName := application.GetName(ctx)
	harborName := n.harbor.Name

	return []*corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      n.harbor.NormalizeComponentName(NotaryServerName),
				Namespace: n.harbor.Namespace,
				Labels: map[string]string{
					"app":      NotaryServerName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			BinaryData: map[string][]byte{
				serverConfigKey: serverConfig,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      n.harbor.NormalizeComponentName(NotarySignerName),
				Namespace: n.harbor.Namespace,
				Labels: map[string]string{
					"app":      NotarySignerName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			BinaryData: map[string][]byte{
				signerConfigKey: signerConfig,
			},
		},
	}
}

func (n *Notary) GetConfigMapsCheckSum() string {
	value := fmt.Sprintf("%x\n%x", serverConfig, signerConfig)
	sum := sha256.New().Sum([]byte(value))

	return fmt.Sprintf("%x", sum)
}
