package notary

import (
	"context"
	"crypto/sha256"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func (n *Notary) GetSecrets(ctx context.Context) []*corev1.Secret {
	return []*corev1.Secret{}
}

func (n *Notary) GetSecretsCheckSum() string {
	// TODO get generation of the secret
	value := fmt.Sprintf("%s\n%s", n.harbor.Spec.Components.Notary.Server.DatabaseSecret, n.harbor.Spec.Components.Notary.Signer.DatabaseSecret)
	sum := sha256.New().Sum([]byte(value))

	return fmt.Sprintf("%x", sum)
}
