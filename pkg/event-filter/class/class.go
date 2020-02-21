package class

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
)

type Filter struct {
	ClassName string
}

// Create returns true if the Create event should be processed
func (cf *Filter) Create(e event.CreateEvent) bool {
	return cf.HarborClassAnnotationMatch(e.Meta)
}

// Delete returns true if the Delete event should be processed
func (cf *Filter) Delete(e event.DeleteEvent) bool {
	return cf.HarborClassAnnotationMatch(e.Meta)
}

// Update returns true if the Update event should be processed
func (cf *Filter) Update(e event.UpdateEvent) bool {
	return cf.HarborClassAnnotationMatch(e.MetaOld) || cf.HarborClassAnnotationMatch(e.MetaNew)
}

// Generic returns true if the Generic event should be processed
func (cf *Filter) Generic(e event.GenericEvent) bool {
	return cf.HarborClassAnnotationMatch(e.Meta)
}

func (cf *Filter) HarborClassAnnotationMatch(meta metav1.Object) bool {
	annotations := meta.GetAnnotations()
	value, ok := annotations[containerregistryv1alpha1.HarborClassAnnotation]

	return value == cf.ClassName || (!ok && cf.ClassName == "")
}
