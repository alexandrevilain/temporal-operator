package v1beta1

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reference to a TemporalCluster.
type TemporalClusterReference struct {
	// The name of the TemporalCluster to reference.
	Name string `json:"name,omitempty"`
	// The namespace of the TemporalCluster to reference.
	// Defaults to the namespace of the requested resource if omitted.
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
}

// NamespacedName returns NamespacedName for the referenced TemporalCluster.
// If the namespace is not set, it uses the provided object's namespace.
func (r *TemporalClusterReference) NamespacedName(obj client.Object) types.NamespacedName {
	namespace := r.Namespace
	if namespace == "" {
		namespace = obj.GetNamespace()
	}
	return types.NamespacedName{Namespace: namespace, Name: r.Name}
}
