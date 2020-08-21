// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ingress

import (
	"github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
)

// GenerateIngressCertAnnotations generates the cert-manager related annotations for cert-manager
// identifying the ingress can creating cert for the detected ingress.
func GenerateIngressCertAnnotations(spec v1alpha1.HarborSpec) map[string]string {
	// Add annotations for cert-manager awareness
	annotations := make(map[string]string)
	issuer := spec.CertificateIssuerRef.Name

	// If name is configured
	if len(issuer) > 0 {
		if spec.CertificateIssuerRef.Kind == v1alpha2.ClusterIssuerKind {
			annotations[v1alpha2.IngressClusterIssuerNameAnnotationKey] = issuer
		} else {
			// Treat as default kind: v1alpha2.IssuerKind
			annotations[v1alpha2.IngressIssuerNameAnnotationKey] = issuer
		}
	}

	return annotations
}
