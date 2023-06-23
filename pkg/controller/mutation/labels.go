package mutation

import (
	"github.com/plotly/harbor-operator/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetLabelsMutation(key, value string, kv ...string) resources.Mutable {
	return GetMetaMutation(metav1.Object.GetLabels, metav1.Object.SetLabels, key, value, kv...)
}

func GetTemplateLabelsMutation(key, value string, kv ...string) resources.Mutable {
	return GetTemplateMetaMutation(metav1.Object.GetLabels, metav1.Object.SetLabels, key, value, kv...)
}
