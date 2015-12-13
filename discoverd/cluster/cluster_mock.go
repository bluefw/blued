package cluster

import (
	"github.com/bluefw/blued/discoverd/api"
)

type MockCluster struct {
}

func NewMockCluster() Cluster {
	return &MockCluster{}
}

func (c MockCluster) RegisterService(ss *api.AppService) error {
	return nil
}

func (c MockCluster) UnregisterService(addr string) error {
	return nil
}
