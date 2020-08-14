// Package v1alpha2 contains API Schema definitions for the containerregistry v1alpha2 API group
// +kubebuilder:object:generate=true
// +groupName=goharbor.io
package v1alpha2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "goharbor.io", Version: "v1alpha2"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

//go:generate controller-gen crd:crdVersions="v1" output:artifacts:config="../../../config/crd/bases" paths="./..."
//go:generate controller-gen webhook output:artifacts:config="../../../config/webhook" paths="./..."
//go:generate controller-gen object paths="./..."
