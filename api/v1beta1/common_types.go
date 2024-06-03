package v1beta1

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TemporalReference is a reference to a object.
type TemporalReference struct {
	// The name of the temporal object to reference.
	Name string `json:"name,omitempty"`
	// The namespace of the temporal object to reference.
	// Defaults to the namespace of the requested resource if omitted.
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
}

// NamespacedName returns NamespacedName for the referenced temporal object.
// If the namespace is not set, it uses the provided object's namespace.
func (r *TemporalReference) NamespacedName(obj client.Object) types.NamespacedName {
	namespace := r.Namespace
	if namespace == "" {
		namespace = obj.GetNamespace()
	}
	return types.NamespacedName{Namespace: namespace, Name: r.Name}
}
