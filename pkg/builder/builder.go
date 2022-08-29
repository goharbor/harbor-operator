package builder

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func ControllerManagedBy(m manager.Manager) *Builder {
	crdGetter, err := apiextv1.NewForConfig(m.GetConfig())
	if err != nil {
		log.Panic(err)
	}

	log := m.GetLogger().WithName("builder")

	return &Builder{
		blder:     ctrl.NewControllerManagedBy(m),
		crdGetter: crdGetter,
		log:       log,
	}
}

// Builder builds a Controller and it can own objects later when the crd of the object ready.
type Builder struct {
	blder *builder.Builder

	ctrl             controller.Controller
	crdGetter        apiextv1.CustomResourceDefinitionsGetter
	forObject        client.Object
	log              logr.Logger
	globalPredicates []predicate.Predicate
	tryOwnsInputs    []tryOwnsInput
}

func (blder *Builder) WithLogger(log logr.Logger) *Builder {
	blder.blder.WithLogger(log)
	blder.log = log

	return blder
}

func (blder *Builder) For(object client.Object, opts ...builder.ForOption) *Builder {
	blder.blder.For(object, opts...)
	blder.forObject = object

	return blder
}

func (blder *Builder) Owns(object client.Object, opts ...builder.OwnsOption) *Builder {
	blder.blder.Owns(object, opts...)

	return blder
}

func (blder *Builder) WithEventFilter(p predicate.Predicate) *Builder {
	blder.blder.WithEventFilter(p)
	blder.globalPredicates = append(blder.globalPredicates, p)

	return blder
}

func (blder *Builder) WithOptions(options controller.Options) *Builder {
	blder.blder.WithOptions(options)

	return blder
}

// TryOwns owns the object when the crdDependency ready, otherwise try owns it after the crdDependency ready.
func (blder *Builder) TryOwns(object client.Object, crdDependency string, predicates ...predicate.Predicate) *Builder {
	if isCRDReady(blder.crdGetter, crdDependency) {
		blder.log.Info("Owns the object directly because the CRD is ready", "crd", crdDependency)

		blder.blder.Owns(object, builder.WithPredicates(predicates...))
	} else {
		blder.log.Info("Will try to own the object laster because the CRD is not ready", "crd", crdDependency)

		blder.tryOwnsInputs = append(blder.tryOwnsInputs, tryOwnsInput{
			object:        object,
			crdDependency: crdDependency,
			predicates:    predicates,
		})
	}

	return blder
}

func (blder *Builder) Build(r reconcile.Reconciler) (controller.Controller, error) {
	var w *tryWatcher

	if len(blder.tryOwnsInputs) > 0 {
		blder.log.Info("Some objects not owned because the CRDs are not ready, we will watch the CRDs and own the objects when they are ready")

		w = &tryWatcher{
			crdGetter:        blder.crdGetter,
			forObject:        blder.forObject,
			log:              blder.log,
			globalPredicates: blder.globalPredicates,
			tryOwnsInputs:    blder.tryOwnsInputs,
		}

		src := &source.Kind{Type: &v1.CustomResourceDefinition{}}
		hdler := &handler.Funcs{
			CreateFunc: func(event.CreateEvent, workqueue.RateLimitingInterface) {
				w.TryWatch()
			},
		}

		blder.blder.Watches(src, hdler)
	}

	ctrl, err := blder.blder.Build(r)
	blder.ctrl = ctrl

	// try to watch immediately to avoid the missing events of crd creating during
	// the Controller building
	if err == nil && w != nil {
		w.WithController(ctrl).TryWatch()
	}

	return ctrl, err
}

func (blder *Builder) Complete(r reconcile.Reconciler) error {
	_, err := blder.Build(r)

	return err
}

type tryOwnsInput struct {
	object        client.Object
	predicates    []predicate.Predicate
	crdDependency string
}

type tryWatcher struct {
	sync.Mutex
	watched map[client.Object]bool

	ctrl      controller.Controller
	crdGetter apiextv1.CustomResourceDefinitionsGetter
	log       logr.Logger

	forObject        client.Object
	globalPredicates []predicate.Predicate
	tryOwnsInputs    []tryOwnsInput
}

func (w *tryWatcher) WithController(ctrl controller.Controller) *tryWatcher {
	w.ctrl = ctrl

	return w
}

func (w *tryWatcher) TryWatch() {
	w.Lock()
	defer w.Unlock()

	if w.ctrl == nil {
		return
	}

	if w.watched == nil {
		w.watched = make(map[client.Object]bool, len(w.tryOwnsInputs))
	}

	for _, own := range w.tryOwnsInputs {
		if w.watched[own.object] || !isCRDReady(w.crdGetter, own.crdDependency) {
			continue
		}

		src := &source.Kind{Type: own.object}
		hdler := &handler.EnqueueRequestForOwner{
			OwnerType:    w.forObject,
			IsController: true,
		}

		allPredicates := append([]predicate.Predicate(nil), w.globalPredicates...)
		allPredicates = append(allPredicates, own.predicates...)

		if err := w.ctrl.Watch(src, hdler, allPredicates...); err != nil {
			w.log.Error(err, "Watch Source Failed", "crd", own.crdDependency)
		} else {
			w.log.Info("Watch Source Success", "crd", own.crdDependency)
			w.watched[own.object] = true
		}
	}
}

func isCRDReady(getter apiextv1.CustomResourceDefinitionsGetter, name string) bool {
	log := ctrl.Log.WithName("builder")

	crdReadyWaitInterval := time.Second * 10 //nolint:gomnd
	crdReadyWaitTimeout := time.Minute * 5   //nolint:gomnd

	// retry to checking the CRD every 10 seconds when it found but not established during 5 minutes
	err := wait.PollImmediate(crdReadyWaitInterval, crdReadyWaitTimeout, func() (bool, error) {
		c, err := getter.CustomResourceDefinitions().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err // not found, poll will return immediately
		}

		for _, cond := range c.Status.Conditions {
			switch cond.Type { //nolint:exhaustive
			case v1.Established:
				if cond.Status == v1.ConditionTrue {
					return true, nil // crd is established，poll will return immediately
				}
			case v1.NamesAccepted:
				if cond.Status == v1.ConditionFalse {
					return false, errors.Errorf("name conflict: %v", cond.Reason) // conflicted, poll will return immediately
				}
			}
		}

		log.Info("CRD found but not established, retry to check again after 10 seconds", "crd", name)

		return false, nil
	})

	return err == nil
}
