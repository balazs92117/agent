package elasticbeanstalk

import (
	rac "github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/resource_attribute_config"
	"github.com/grafana/river"
)

const Name = "elasticbeanstalk"

type Config struct {
	ResourceAttributes ResourceAttributesConfig `river:"resource_attributes,block,optional"`
}

// DefaultArguments holds default settings for Config.
var DefaultArguments = Config{
	ResourceAttributes: ResourceAttributesConfig{
		CloudPlatform:         rac.ResourceAttributeConfig{Enabled: true},
		CloudProvider:         rac.ResourceAttributeConfig{Enabled: true},
		DeploymentEnvironment: rac.ResourceAttributeConfig{Enabled: true},
		ServiceInstanceID:     rac.ResourceAttributeConfig{Enabled: true},
		ServiceVersion:        rac.ResourceAttributeConfig{Enabled: true},
	},
}

var _ river.Defaulter = (*Config)(nil)

// SetToDefault implements river.Defaulter.
func (args *Config) SetToDefault() {
	*args = DefaultArguments
}

func (args Config) Convert() map[string]interface{} {
	return map[string]interface{}{
		"resource_attributes": args.ResourceAttributes.Convert(),
	}
}

// ResourceAttributesConfig provides config for elastic_beanstalk resource attributes.
type ResourceAttributesConfig struct {
	CloudPlatform         rac.ResourceAttributeConfig `river:"cloud.platform,block,optional"`
	CloudProvider         rac.ResourceAttributeConfig `river:"cloud.provider,block,optional"`
	DeploymentEnvironment rac.ResourceAttributeConfig `river:"deployment.environment,block,optional"`
	ServiceInstanceID     rac.ResourceAttributeConfig `river:"service.instance.id,block,optional"`
	ServiceVersion        rac.ResourceAttributeConfig `river:"service.version,block,optional"`
}

func (r ResourceAttributesConfig) Convert() map[string]interface{} {
	return map[string]interface{}{
		"cloud.platform":         r.CloudPlatform.Convert(),
		"cloud.provider":         r.CloudProvider.Convert(),
		"deployment.environment": r.DeploymentEnvironment.Convert(),
		"service.instance.id":    r.ServiceInstanceID.Convert(),
		"service.version":        r.ServiceVersion.Convert(),
	}
}
