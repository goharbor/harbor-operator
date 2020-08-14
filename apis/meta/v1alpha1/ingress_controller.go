package v1alpha1

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum={"default","gce","ncp"}
// Type of ingress controller if it has specific requirements.
type IngressController string

const (
	// Default ingress controller
	IngressControllerDefault IngressController = "default"
	// Google Cloud Engine ingress controller
	IngressControllerGCE IngressController = "gce"
	// NSX-T Container Plugin ingress controller
	IngressControllerNCP IngressController = "ncp"
)
