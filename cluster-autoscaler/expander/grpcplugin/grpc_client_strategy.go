package grpcplugin

import (
  "k8s.io/autoscaler/cluster-autoscaler/expander"
  schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type grpcclientstrategy struct {
  fallbackStrategy expander.Strategy
  grpcClient expanderClient
}

// NewStrategy returns a strategy that calls out to an external gRPC server to find the best option
func NewStrategy(fallbackStrategy expander.Strategy) expander.Strategy {
  return &grpcclientstrategy{fallbackStrategy: fallbackStrategy}
}

func (g *grpcclientstrategy) BestOption(expansionOptions []expander.Option, nodeInfo map[string]*schedulerframework.NodeInfo) *expander.Option {

}