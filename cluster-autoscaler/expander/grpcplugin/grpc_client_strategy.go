package grpcplugin

import (
  "log"
  "context"
  "time"

  v1 "k8s.io/api/core/v1"
  "k8s.io/autoscaler/cluster-autoscaler/expander"
  schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials"
)

type grpcclientstrategy struct {
  fallbackStrategy expander.Strategy
  grpcClient ExpanderClient
}

// NewStrategy returns a strategy that calls out to an external gRPC server to find the best option
func NewStrategy(fallbackStrategy expander.Strategy) expander.Strategy {
  // ca file
  // if ca file
  caFile := "x50a/ca_cert.pem"
  creds, err := credentials.NewClientTLSFromFile(*caFile, "serverHostnameOverride")
  if err != nil {
    log.Fatalf("Failed to create TLS credentials %v", err)
  }
  dialOpt := grpc.WithTransportCredentials(creds)
  // else, insexcure
  dialOpt = grpc.WithInsecure()
  conn, err := grpc.Dial("serveraddr", dialOpt)
  if err != nil {
    log.Fatalf("fail to dial: %v", err)
  }
  defer conn.Close()
  client := NewExpanderClient(conn)
  return &grpcclientstrategy{fallbackStrategy: fallbackStrategy, grpcClient: client}
}

func (g *grpcclientstrategy) BestOption(expansionOptions []expander.Option, nodeInfo map[string]*schedulerframework.NodeInfo) *expander.Option {
  // Transform inputs to gRPC inputs
  nodeGroupIDOptionMap := make(map[string]expander.Option)
  grpcOptionsSlice := []*Option{}
  populateOptionsForGRPC(expansionOptions, nodeGroupIDOptionMap, &grpcOptionsSlice)
  grpcNodeInfoMap := make(map[string]*v1.Node)
  populateNodeInfoForGRPC(nodeInfo, grpcNodeInfoMap)

  // call gRPC server to get BestOption
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()
  bestOptionResponse, err := g.grpcClient.BestOption(ctx, &BestOptionRequest{Options: grpcOptionsSlice, NodeInfoMap: grpcNodeInfoMap})
  if err != nil {
    return g.fallbackStrategy.BestOption(expansionOptions, nodeInfo)
  }

  // Transform back to bestOption
  bestOption := nodeGroupIDOptionMap[bestOptionResponse.Option.NodeGroupId]
  return &bestOption
}

// populateOptionsForGRPC creates a map of nodegroup ID and options, as well as a slice of Options objects for the gRPC call
func populateOptionsForGRPC(expansionOptions []expander.Option, nodeGroupIDOptionMap map[string]expander.Option, grpcOptionsSlice *[]*Option) {
  for _, option := range expansionOptions {
    nodeGroupIDOptionMap[option.NodeGroup.Id()] = option
    *grpcOptionsSlice = append(*grpcOptionsSlice, newOptionMessage(option.NodeGroup.Id(), int32(option.NodeCount), option.Debug, option.Pods))
  }
}

// populateNodeInfoForGRPC modifies the nodeInfo object, and replaces it with the v1.Node to pass through grpc
func populateNodeInfoForGRPC(nodeInfos map[string]*schedulerframework.NodeInfo, grpcNodeInfoMap map[string]*v1.Node) {
  for nodeId, nodeInfo := range nodeInfos {
    grpcNodeInfoMap[nodeId] = nodeInfo.Node()
  }
}

func newOptionMessage(nodeGroupId string, nodeCount int32, debug string, pods []*v1.Pod) *Option{
  return &Option{NodeGroupId: nodeGroupId, NodeCount: nodeCount, Debug: debug, Pod: pods}
}
