package provider

import (
	"fmt"

	"github.com/huynt0812/dbaas-operator/pkg/provider/cnpg"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DefaultFactory is the default provider factory implementation
type DefaultFactory struct{}

// GetProvider returns a provider instance for the given engine type
func (f *DefaultFactory) GetProvider(engineType string, c client.Client, scheme *runtime.Scheme) (Provider, error) {
	switch engineType {
	case "postgresql":
		return cnpg.NewProvider(c, scheme), nil
	case "mongodb":
		return nil, fmt.Errorf("MongoDB provider not implemented yet")
	case "mysql":
		return nil, fmt.Errorf("MySQL provider not implemented yet")
	case "kafka":
		return nil, fmt.Errorf("Kafka provider not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported engine type: %s", engineType)
	}
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() ProviderFactory {
	return &DefaultFactory{}
}
