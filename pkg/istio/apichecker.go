package istio

import (
	"context"
	"fmt"

	istioapinetworkingv1beta1 "istio.io/api/networking/v1beta1"
	istioapisecurityv1beta1 "istio.io/api/security/v1beta1"
	istioapiv1beta1 "istio.io/api/type/v1beta1"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var fakeObjects = []client.Object{
	&istionetworkingv1beta1.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "fake-",
		},
		Spec: istioapinetworkingv1beta1.DestinationRule{
			Host: "fake",
			TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
				Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
					Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
				},
			},
		},
	},
	&istiosecurityv1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "fake-",
		},
		Spec: istioapisecurityv1beta1.PeerAuthentication{
			Selector: &istioapiv1beta1.WorkloadSelector{
				MatchLabels: map[string]string{
					"fake": "fake",
				},
			},
			Mtls: &istioapisecurityv1beta1.PeerAuthentication_MutualTLS{
				Mode: istioapisecurityv1beta1.PeerAuthentication_MutualTLS_STRICT,
			},
		},
	},
}

// APIChecker checks for needed istio resources in the cluster.
// Under-the-hood is uses a dry-run client to check if CRDs are available.
// APIChecker is inspired by cert-manager's apichecker
// (https://github.com/cert-manager/cert-manager/blob/master/pkg/util/cmapichecker).
type APIChecker struct {
	client client.Client
}

// NewAPIChecker creates a new ApiChecker.
func NewAPIChecker(restcfg *rest.Config, scheme *runtime.Scheme, namespace string) (*APIChecker, error) {
	if err := istiosecurityv1beta1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("can't add istiosecurityv1beta1 to scheme: %w", err)
	}
	if err := istionetworkingv1beta1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("can't add istionetworkingv1beta1 to scheme: %w", err)
	}

	cl, err := client.New(restcfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("can't create client: %w", err)
	}

	return &APIChecker{
		client: client.NewNamespacedClient(client.NewDryRunClient(cl), namespace),
	}, nil
}

// Check attempts to perform a dry-run create of a needed istio resources.
func (c *APIChecker) Check(ctx context.Context) error {
	for _, fakeObject := range fakeObjects {
		err := c.client.Create(ctx, fakeObject)
		if err != nil {
			return err
		}
	}
	return nil
}
