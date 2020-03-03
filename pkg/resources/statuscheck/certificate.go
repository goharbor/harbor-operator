package statuscheck

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CertificateCheck(ctx context.Context, object runtime.Object) (bool, error) {
	cert := object.(*certv1.Certificate)

	expiration := cert.Status.NotAfter
	if !expiration.IsZero() && metav1.Now().After(expiration.Time) {
		// Certificate expired
		return false, nil
	}

	for _, condition := range cert.Status.Conditions {
		if condition.Type == certv1.CertificateConditionReady {
			return condition.Status == cmmeta.ConditionTrue, nil
		}
	}

	return false, nil
}
