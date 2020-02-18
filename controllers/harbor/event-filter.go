package harbor

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/event"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/pkg/errors"
)

type EventFilter struct {
	ClassName string
	Scheme    *runtime.Scheme
}

// Create returns true if the Create event should be processed
func (ef *EventFilter) Create(e event.CreateEvent) bool {
	return ef.HarborClassAnnotationMatch(e.Meta) || ef.IsOwned(e.Meta, e.Object)
}

// Delete returns true if the Delete event should be processed
func (ef *EventFilter) Delete(e event.DeleteEvent) bool {
	return ef.HarborClassAnnotationMatch(e.Meta) || ef.IsOwned(e.Meta, e.Object)
}

// Update returns true if the Update event should be processed
func (ef *EventFilter) Update(e event.UpdateEvent) bool {
	return (ef.HarborClassAnnotationMatch(e.MetaOld) || ef.IsOwned(e.MetaOld, e.ObjectOld)) ||
		(ef.HarborClassAnnotationMatch(e.MetaNew) || ef.IsOwned(e.MetaNew, e.ObjectNew))
}

// Generic returns true if the Generic event should be processed
func (ef *EventFilter) Generic(e event.GenericEvent) bool {
	return ef.HarborClassAnnotationMatch(e.Meta) || ef.IsOwned(e.Meta, e.Object)
}

func (ef *EventFilter) HarborClassAnnotationMatch(meta metav1.Object) bool {
	annotations := meta.GetAnnotations()
	value, ok := annotations[containerregistryv1alpha1.HarborClassAnnotation]

	return value == ef.ClassName || (!ok && ef.ClassName == "")
}

func (ef *EventFilter) IsOwned(meta metav1.Object, ro runtime.Object) bool {
	gvk, err := apiutil.GVKForObject(ro, ef.Scheme)
	if err != nil {
		panic(errors.Wrap(err, "cannot get group version kind"))
	}

	owners := meta.GetOwnerReferences()
	for _, owner := range owners {
		if owner.Controller != nil && *owner.Controller {
			return owner.Kind == gvk.Kind
		}
	}

	return false
}

func (r *Reconciler) GetEventFilter() *EventFilter {
	return &EventFilter{
		ClassName: r.Config.ClassName,
		Scheme:    r.Scheme,
	}
}
