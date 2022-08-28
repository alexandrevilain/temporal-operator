package istio_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alexandrevilain/temporal-operator/pkg/istio"
	"github.com/stretchr/testify/assert"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type fakeErrorClient struct {
	client.Client

	createError error
}

func (cl *fakeErrorClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if cl.createError != nil {
		return cl.createError
	}

	return cl.Client.Create(ctx, obj, opts...)
}

var (
	errNoKindDestinationrule = `error finding the scope of the object: failed to get restmapping: no matches for kind "DestinationRule" in group "networking.istio.io"`
)

func TestApiChecker(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(istiosecurityv1beta1.AddToScheme(scheme))
	utilruntime.Must(istionetworkingv1beta1.AddToScheme(scheme))
	clientWithScheme := fake.NewClientBuilder().WithScheme(scheme).Build()

	tests := map[string]struct {
		client        client.Client
		expectedError string
	}{
		"resource not registered in scheme": {
			client:        fake.NewClientBuilder().Build(),
			expectedError: "no kind is registered for the type v1beta1.DestinationRule in scheme",
		},
		"api server returning error": {
			client: &fakeErrorClient{
				Client:      clientWithScheme,
				createError: errors.New(errNoKindDestinationrule),
			},
			expectedError: errNoKindDestinationrule,
		},
		"scheme and api allowing request": {
			client:        clientWithScheme,
			expectedError: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			c := istio.NewAPICheckerForTesting(test.client)

			err := c.Check(context.Background())
			if test.expectedError != "" {
				assert.Error(tt, err)
				assert.Contains(tt, err.Error(), test.expectedError)
			} else {
				assert.NoError(tt, err)
			}
		})
	}
}
