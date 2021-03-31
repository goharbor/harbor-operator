package checksum

import (
	"encoding/json"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func Test(t *testing.T) {
	resource := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "resource",
			Namespace: "namespace",
		},
		Spec: appsv1.DeploymentSpec{
			Paused: false,
		},
	}

	unstructuredObj, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	fmt.Print(tojson(unstructured.Unstructured{unstructuredObj}, "spec.template"))

}

func tojson(object unstructured.Unstructured, filed ...string) string {
	obj, _, _ := unstructured.NestedFieldCopy(object.Object, filed[0])
	str, _ := json.Marshal(obj)
	return string(str)
}
