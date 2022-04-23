package resource

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ServiceHeadlessService = "headless-service"
	ServiceConfig          = "config"
	ServiceDeployment      = "deployment"
)

type Builder interface {
	Build() (client.Object, error)
	Update(client.Object) error
}
