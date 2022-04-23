package cluster

import (
	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"go.temporal.io/server/common"
	"k8s.io/apimachinery/pkg/runtime"
)

type TemporalClusterBuilder struct {
	Instance *v1alpha1.TemporalCluster
	Scheme   *runtime.Scheme
}

func (b *TemporalClusterBuilder) ResourceBuilders() ([]resource.Builder, error) {
	builders := []resource.Builder{
		resource.NewConfigmapBuilder(b.Instance, b.Scheme),
		resource.NewFrontendServiceBuilder(b.Instance, b.Scheme),
	}

	for _, serviceName := range []string{
		common.FrontendServiceName,
		common.HistoryServiceName,
		common.MatchingServiceName,
		common.WorkerServiceName,
	} {
		specs, err := b.Instance.Spec.Services.GetServiceSpec(serviceName)
		if err != nil {
			return nil, err
		}
		builders = append(builders, resource.NewDeploymentBuilder(serviceName, b.Instance, b.Scheme, specs))
		builders = append(builders, resource.NewHeadlessServiceBuilder(serviceName, b.Instance, b.Scheme, specs))
	}

	return builders, nil
}
