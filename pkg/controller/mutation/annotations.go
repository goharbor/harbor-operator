package mutation

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/goharbor/harbor-operator/pkg/resources"
)

func GetAnnotationsMutation(key, value string, kv ...string) resources.Mutable {
	return GetMetaMutation(metav1.Object.GetAnnotations, metav1.Object.SetAnnotations, key, value, kv...)
}

func GetTemplateAnnotationsMutation(key, value string, kv ...string) resources.Mutable {
	return GetTemplateMetaMutation(metav1.Object.GetAnnotations, metav1.Object.SetAnnotations, key, value, kv...)
}
