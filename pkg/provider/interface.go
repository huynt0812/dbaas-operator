package provider

import (
	"github.com/huynt0812/dbaas-operator/pkg/provider/interfaces"
)

// Re-export interfaces for backward compatibility
type Provider = interfaces.Provider
type Applier = interfaces.Applier
type OperationsHandler = interfaces.OperationsHandler
type ProviderFactory = interfaces.ProviderFactory
