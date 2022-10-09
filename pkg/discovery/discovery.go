package discovery

import (
	"github.com/vumanhcuongit/scan/internal/config"
)

//go:generate mockgen -package=discovery -destination=discovery_mock.go -source=discovery.go
type IServiceDiscovery interface {
}

type ServiceDiscovery struct {
}

func (s *ServiceDiscovery) Close() error {
	return nil
}

func NewServiceDiscovery(cfg *config.App) (*ServiceDiscovery, error) {
	return &ServiceDiscovery{}, nil
}
