package v1alpha1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/webhook/resourcesemantics"
)

var _ resourcesemantics.GenericCRD = (*Function)(nil)

const DefaultingConfigKey = "defaulting-config"

type DefaultingConfig struct {
	RequestCpu    string `envconfig:"default=50m"`
	RequestMemory string `envconfig:"default=64Mi"`
	LimitsCpu     string `envconfig:"default=100m"`
	LimitsMemory  string `envconfig:"default=128Mi"`
	MinReplicas   int32  `envconfig:"default=1"`
	MaxReplicas   int32  `envconfig:"default=1"`
}

func (fn *Function) SetDefaults(ctx context.Context) {
	fn.Spec.defaultResources(ctx)
	fn.Spec.defaultReplicas(ctx)
}

func (spec *FunctionSpec) defaultReplicas(ctx context.Context) {

	if spec.MinReplicas == nil {
		newMin := ctx.Value(DefaultingConfigKey).(DefaultingConfig).MinReplicas
		if spec.MaxReplicas != nil && *spec.MaxReplicas < newMin {
			newMin = *spec.MaxReplicas
		}

		spec.MinReplicas = &newMin
	}
	if spec.MaxReplicas == nil {
		newMax := ctx.Value(DefaultingConfigKey).(DefaultingConfig).MaxReplicas
		if *spec.MinReplicas > newMax {
			newMax = *spec.MinReplicas
		}

		spec.MaxReplicas = &newMax
	}
}

func (spec *FunctionSpec) defaultResources(ctx context.Context) {
	resources := spec.Resources
	if resources.Requests == nil {
		spec.Resources.Requests = corev1.ResourceList{}
	}
	if resources.Requests.Memory().IsZero() {
		val := ctx.Value(DefaultingConfigKey).(DefaultingConfig).RequestMemory
		newRerource := resource.MustParse(val)
		if !resources.Limits.Memory().IsZero() && resources.Limits.Memory().Cmp(newRerource) == -1 {
			newRerource = *resources.Limits.Memory()
		}

		spec.Resources.Requests[corev1.ResourceMemory] = newRerource
	}
	if resources.Requests.Cpu().IsZero() {
		newRerource := resource.MustParse(ctx.Value(DefaultingConfigKey).(DefaultingConfig).RequestCpu)
		if !resources.Limits.Cpu().IsZero() && resources.Limits.Cpu().Cmp(newRerource) == -1 {
			newRerource = *resources.Limits.Cpu()
		}

		spec.Resources.Requests[corev1.ResourceCPU] = newRerource
	}

	if resources.Limits == nil {
		spec.Resources.Limits = corev1.ResourceList{}
	}
	if resources.Limits.Memory().IsZero() {
		newRerource := resource.MustParse(ctx.Value(DefaultingConfigKey).(DefaultingConfig).LimitsMemory)
		if spec.Resources.Requests.Memory().Cmp(newRerource) == 1 {
			newRerource = *spec.Resources.Requests.Memory()
		}

		spec.Resources.Limits[corev1.ResourceMemory] = newRerource
	}
	if resources.Limits.Cpu().IsZero() {
		newRerource := resource.MustParse(ctx.Value(DefaultingConfigKey).(DefaultingConfig).LimitsCpu)
		if spec.Resources.Requests.Cpu().Cmp(newRerource) == 1 {
			newRerource = *spec.Resources.Requests.Cpu()
		}

		spec.Resources.Limits[corev1.ResourceCPU] = newRerource
	}
}
