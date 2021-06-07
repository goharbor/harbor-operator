package class

import (
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type Filter struct {
	ClassName string
}

// Create returns true if the Create event should be processed.
func (cf *Filter) Create(e event.CreateEvent) bool {
	return cf.HarborClassAnnotationMatch(e.Object)
}

// Delete returns true if the Delete event should be processed.
func (cf *Filter) Delete(e event.DeleteEvent) bool {
	return cf.HarborClassAnnotationMatch(e.Object)
}

// Update returns true if the Update event should be processed.
func (cf *Filter) Update(e event.UpdateEvent) bool {
	return cf.HarborClassAnnotationMatch(e.ObjectOld) || cf.HarborClassAnnotationMatch(e.ObjectNew)
}

// Generic returns true if the Generic event should be processed.
func (cf *Filter) Generic(e event.GenericEvent) bool {
	return cf.HarborClassAnnotationMatch(e.Object)
}

func (cf *Filter) HarborClassAnnotationMatch(meta metav1.Object) bool {
	annotations := meta.GetAnnotations()
	value, ok := annotations[goharborv1.HarborClassAnnotation]

	return value == cf.ClassName || (!ok && cf.ClassName == "")
}
