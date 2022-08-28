package istio

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewAPICheckerForTesting(client client.Client) *APIChecker {
	return &APIChecker{
		client: client,
	}
}
