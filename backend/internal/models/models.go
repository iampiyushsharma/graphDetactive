package models

import "time"

// Position represents 2D coordinates for React Flow positioning
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NodeData represents properties embedded in a React Flow node
type NodeData struct {
	Label      string         `json:"label"`
	Type       string         `json:"type,omitempty"`
	Properties map[string]any `json:"properties"`
}

// RFNode represents a node formatted for React Flow
type RFNode struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"` // Custom React Flow component type
	Data     NodeData `json:"data"`
	Position Position `json:"position"`
}

// RFEdge represents a directed relationship formatted for React Flow
type RFEdge struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	Label    string `json:"label,omitempty"`
	Animated bool   `json:"animated"`
}

// GraphResponse is returned by graph-native endpoints to render directly in React Flow
type GraphResponse struct {
	Nodes []RFNode `json:"nodes"`
	Edges []RFEdge `json:"edges"`
}

// Incident holds structured data for list and detail representations
type Incident struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	ServiceName string    `json:"serviceName,omitempty"`
}

// Service holds metadata for services and databases
type Service struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Language string `json:"language,omitempty"`
	Status   string `json:"status"`
}
