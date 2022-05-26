package status

import "github.com/alexandrevilain/temporal-operator/api/v1alpha1"

// AddServiceStatus adds the provided service status to the cluster's status.
func AddServiceStatus(c *v1alpha1.TemporalCluster, status *v1alpha1.ServiceStatus) {
	found := false
	for _, serviceStatus := range c.Status.Services {
		if serviceStatus.Name == status.Name {
			found = true
			serviceStatus.Version = status.Version
		}
	}
	if !found {
		c.Status.Services = append(c.Status.Services, *status)
	}
}
