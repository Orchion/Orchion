package router

// Message represents a message to be routed between nodes
type Message struct {
	SourceNodeID string
	TargetNodeID string
	Payload      []byte
}

// Router is responsible for routing messages between nodes
type Router interface {
	// RouteMessage routes a message from source to target node
	RouteMessage(msg *Message) error
	
	// GetRoutingStats returns statistics about routing decisions
	GetRoutingStats() *RoutingStats
}

// RoutingStats contains statistics about routing decisions
type RoutingStats struct {
	TotalRouted   int64
	FailedRoutes  int64
}
