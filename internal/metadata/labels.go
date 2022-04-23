package metadata

// LabelsSelector returns service's default labels.
func LabelsSelector(clusterName, serviceName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      clusterName,
		"app.kubernetes.io/component": serviceName,
		"app.kubernetes.io/part-of":   "temporal",
	}
}

// GetLabels returns a Labels for a temporal service.
func GetLabels(name, service string, labels map[string]string) map[string]string {
	l := LabelsSelector(name, service)
	for k, v := range labels {
		l[k] = v
	}
	return l
}
