package istio_test

import (
	"testing"

	"github.com/alexandrevilain/temporal-operator/pkg/resource/istio"
	"github.com/stretchr/testify/assert"
	istioapinetworkingv1beta1 "istio.io/api/networking/v1beta1"
)

func TestDestinationRuleEqual(t *testing.T) {
	tests := map[string]struct {
		a     istioapinetworkingv1beta1.DestinationRule
		b     istioapinetworkingv1beta1.DestinationRule
		equal bool
	}{
		"equal": {
			a: istioapinetworkingv1beta1.DestinationRule{
				Host: "host",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
			},
			b: istioapinetworkingv1beta1.DestinationRule{
				Host: "host",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
			},
			equal: true,
		},
		"host differs": {
			a: istioapinetworkingv1beta1.DestinationRule{
				Host: "host-a",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
			},
			b: istioapinetworkingv1beta1.DestinationRule{
				Host: "host-b",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
			},
			equal: false,
		},
		"traffic policy differs": {
			a: istioapinetworkingv1beta1.DestinationRule{
				Host: "host",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
			},
			b: istioapinetworkingv1beta1.DestinationRule{
				Host: "host",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_MUTUAL,
					},
				},
			},
			equal: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			b := &istio.DestinationRuleBuilder{}
			result := b.Equal(test.a, test.b)
			assert.Equal(tt, test.equal, result)
		})
	}
}
