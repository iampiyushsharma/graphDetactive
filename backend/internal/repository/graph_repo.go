package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/novacart/graph-detective/internal/database"
	"github.com/novacart/graph-detective/internal/models"
)

type GraphRepository struct {
	db *database.DB
}

func NewGraphRepository(db *database.DB) *GraphRepository {
	return &GraphRepository{db: db}
}

// Helper to parse time from Neo4j property
func parseTime(val any) time.Time {
	if t, ok := val.(time.Time); ok {
		return t
	}
	if s, ok := val.(string); ok {
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

// GetIncidents returns all active incidents with the name of the service they affect
func (r *GraphRepository) GetIncidents(ctx context.Context) ([]models.Incident, error) {
	if r.db.Driver == nil {
		return nil, fmt.Errorf("database driver is not initialized")
	}

	session := r.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (i:Incident)-[:AFFECTS]->(s)
		RETURN i, s.name AS serviceName
		ORDER BY i.createdAt DESC
	`

	res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		var incidents []models.Incident
		for result.Next(ctx) {
			record := result.Record()
			iNodeObj, found := record.Get("i")
			if !found {
				continue
			}

			iNode := iNodeObj.(neo4j.Node)
			props := iNode.Props

			var serviceName string
			if sNameVal, ok := record.Get("serviceName"); ok && sNameVal != nil {
				serviceName = sNameVal.(string)
			}

			id, _ := props["id"].(string)
			title, _ := props["title"].(string)
			severity, _ := props["severity"].(string)
			status, _ := props["status"].(string)
			description, _ := props["description"].(string)

			incidents = append(incidents, models.Incident{
				ID:          id,
				Title:       title,
				Severity:    severity,
				Status:      status,
				Description: description,
				CreatedAt:   parseTime(props["createdAt"]),
				ServiceName: serviceName,
			})
		}
		return incidents, nil
	})

	if err != nil {
		return nil, err
	}
	return res.([]models.Incident), nil
}

// GetIncidentByID fetches details of a specific incident
func (r *GraphRepository) GetIncidentByID(ctx context.Context, id string) (*models.Incident, error) {
	if r.db.Driver == nil {
		return nil, fmt.Errorf("database driver is not initialized")
	}

	session := r.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (i:Incident {id: $id})-[:AFFECTS]->(s)
		RETURN i, s.name AS serviceName
	`

	res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			iNodeObj, _ := record.Get("i")
			iNode := iNodeObj.(neo4j.Node)
			props := iNode.Props

			var serviceName string
			if sNameVal, ok := record.Get("serviceName"); ok && sNameVal != nil {
				serviceName = sNameVal.(string)
			}

			idVal, _ := props["id"].(string)
			title, _ := props["title"].(string)
			severity, _ := props["severity"].(string)
			status, _ := props["status"].(string)
			description, _ := props["description"].(string)

			return &models.Incident{
				ID:          idVal,
				Title:       title,
				Severity:    severity,
				Status:      status,
				Description: description,
				CreatedAt:   parseTime(props["createdAt"]),
				ServiceName: serviceName,
			}, nil
		}
		return nil, fmt.Errorf("incident with ID %s not found", id)
	})

	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("incident not found")
	}
	return res.(*models.Incident), nil
}

// GetRootCause traces the graph from the incident to find causal paths
func (r *GraphRepository) GetRootCause(ctx context.Context, incidentID string) (*models.GraphResponse, error) {
	if r.db.Driver == nil {
		return nil, fmt.Errorf("database driver is not initialized")
	}

	session := r.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// Cypher query to trace: Incident -> Service -> Deployment -> Commit -> ConfigChange
	query := `
		MATCH path = (i:Incident {id: $incidentId})-[:AFFECTS]->(s)
					  <-[:TO_SERVICE|TO_DATABASE]-(d:Deployment)
					  -[:CONTAINS]->(c:Commit)
					  -[:MODIFIED_CONFIG]->(cc:ConfigChange)
		RETURN path
	`

	res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, map[string]any{"incidentId": incidentID})
		if err != nil {
			return nil, err
		}

		return r.parsePaths(ctx, result)
	})

	if err != nil {
		return nil, err
	}
	return res.(*models.GraphResponse), nil
}

// GetBlastRadius traverses downstream services that depend on a selected component
func (r *GraphRepository) GetBlastRadius(ctx context.Context, componentID string) (*models.GraphResponse, error) {
	if r.db.Driver == nil {
		return nil, fmt.Errorf("database driver is not initialized")
	}

	session := r.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// Traverses variable depth (1 to 5 hops) downstream dependencies
	query := `
		MATCH path = (downstream:Service)-[:DEPENDS_ON*1..5]->(target {id: $componentId})
		RETURN path
	`

	res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, map[string]any{"componentId": componentID})
		if err != nil {
			return nil, err
		}

		return r.parsePaths(ctx, result)
	})

	if err != nil {
		return nil, err
	}
	return res.(*models.GraphResponse), nil
}

// GetDependencies returns the direct (1-hop) upstream and downstream connections for a service
func (r *GraphRepository) GetDependencies(ctx context.Context, serviceID string) (*models.GraphResponse, error) {
	if r.db.Driver == nil {
		return nil, fmt.Errorf("database driver is not initialized")
	}

	session := r.db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (s {id: $serviceId})
		OPTIONAL MATCH (s)-[r1:DEPENDS_ON]->(u)
		OPTIONAL MATCH (d)-[r2:DEPENDS_ON]->(s)
		RETURN s, collect(distinct u) as upstreams, collect(distinct r1) as r1s, collect(distinct d) as downstreams, collect(distinct r2) as r2s
	`

	res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, map[string]any{"serviceId": serviceID})
		if err != nil {
			return nil, err
		}

		nodesMap := make(map[string]models.RFNode)
		edgesMap := make(map[string]models.RFEdge)
		elementToBusinessID := make(map[string]string)

		if result.Next(ctx) {
			record := result.Record()

			// 1. Process source node
			sNodeObj, _ := record.Get("s")
			if sNodeObj != nil {
				sNode := sNodeObj.(neo4j.Node)
				rfNode := parseRFNode(sNode)
				nodesMap[rfNode.ID] = rfNode
				elementToBusinessID[sNode.ElementId] = rfNode.ID
			}

			// 2. Process upstreams
			upstreamsObj, _ := record.Get("upstreams")
			if upstreamsObj != nil {
				for _, uVal := range upstreamsObj.([]any) {
					if uVal == nil {
						continue
					}
					uNode := uVal.(neo4j.Node)
					rfNode := parseRFNode(uNode)
					nodesMap[rfNode.ID] = rfNode
					elementToBusinessID[uNode.ElementId] = rfNode.ID
				}
			}

			// 3. Process downstreams
			downstreamsObj, _ := record.Get("downstreams")
			if downstreamsObj != nil {
				for _, dVal := range downstreamsObj.([]any) {
					if dVal == nil {
						continue
					}
					dNode := dVal.(neo4j.Node)
					rfNode := parseRFNode(dNode)
					nodesMap[rfNode.ID] = rfNode
					elementToBusinessID[dNode.ElementId] = rfNode.ID
				}
			}

			// 4. Process upstream relationships (r1s)
			r1sObj, _ := record.Get("r1s")
			if r1sObj != nil {
				for _, r1Val := range r1sObj.([]any) {
					if r1Val == nil {
						continue
					}
					rel := r1Val.(neo4j.Relationship)
					edge := parseRFEdge(rel)
					edgesMap[edge.ID] = edge
				}
			}

			// 5. Process downstream relationships (r2s)
			r2sObj, _ := record.Get("r2s")
			if r2sObj != nil {
				for _, r2Val := range r2sObj.([]any) {
					if r2Val == nil {
						continue
					}
					rel := r2Val.(neo4j.Relationship)
					edge := parseRFEdge(rel)
					edgesMap[edge.ID] = edge
				}
			}
		}

		// Map element IDs in edges to business IDs
		finalEdges := []models.RFEdge{}
		for _, edge := range edgesMap {
			srcBusiness, existsSrc := elementToBusinessID[edge.Source]
			if existsSrc {
				edge.Source = srcBusiness
			}
			tgtBusiness, existsTgt := elementToBusinessID[edge.Target]
			if existsTgt {
				edge.Target = tgtBusiness
			}
			finalEdges = append(finalEdges, edge)
		}

		finalNodes := []models.RFNode{}
		for _, node := range nodesMap {
			finalNodes = append(finalNodes, node)
		}

		return &models.GraphResponse{
			Nodes: finalNodes,
			Edges: finalEdges,
		}, nil
	})

	if err != nil {
		return nil, err
	}
	return res.(*models.GraphResponse), nil
}

// parsePaths processes Neo4j records containing "path" variables and constructs nodes/edges lists
func (r *GraphRepository) parsePaths(ctx context.Context, result neo4j.ResultWithContext) (*models.GraphResponse, error) {
	nodesMap := make(map[string]models.RFNode)
	edgesMap := make(map[string]models.RFEdge)
	elementToBusinessID := make(map[string]string)

	for result.Next(ctx) {
		record := result.Record()
		pathObj, found := record.Get("path")
		if !found || pathObj == nil {
			continue
		}

		path := pathObj.(neo4j.Path)

		// Extract nodes
		for _, node := range path.Nodes {
			rfNode := parseRFNode(node)
			nodesMap[rfNode.ID] = rfNode
			elementToBusinessID[node.ElementId] = rfNode.ID
		}

		// Extract relationships
		for _, rel := range path.Relationships {
			edge := parseRFEdge(rel)
			edgesMap[edge.ID] = edge
		}
	}

	// Map element IDs in edges to business IDs
	finalEdges := []models.RFEdge{}
	for _, edge := range edgesMap {
		srcBusiness, existsSrc := elementToBusinessID[edge.Source]
		if existsSrc {
			edge.Source = srcBusiness
		}
		tgtBusiness, existsTgt := elementToBusinessID[edge.Target]
		if existsTgt {
			edge.Target = tgtBusiness
		}
		finalEdges = append(finalEdges, edge)
	}

	finalNodes := []models.RFNode{}
	for _, node := range nodesMap {
		finalNodes = append(finalNodes, node)
	}

	return &models.GraphResponse{
		Nodes: finalNodes,
		Edges: finalEdges,
	}, nil
}

// Helper to convert neo4j.Node into RFNode model
func parseRFNode(node neo4j.Node) models.RFNode {
	props := node.Props
	labels := node.Labels

	nodeType := "unknown"
	if len(labels) > 0 {
		nodeType = labels[0]
	}

	id, ok := props["id"].(string)
	if !ok {
		id = node.ElementId
	}

	label := id
	if nameVal, ok := props["name"].(string); ok {
		label = nameVal
	} else if titleVal, ok := props["title"].(string); ok {
		label = titleVal
	}

	return models.RFNode{
		ID:   id,
		Type: nodeType,
		Data: models.NodeData{
			Label:      label,
			Type:       nodeType,
			Properties: props,
		},
		Position: models.Position{X: 0, Y: 0},
	}
}

// Helper to convert neo4j.Relationship into RFEdge model
func parseRFEdge(rel neo4j.Relationship) models.RFEdge {
	return models.RFEdge{
		ID:       rel.ElementId,
		Source:   rel.StartElementId,
		Target:   rel.EndElementId,
		Label:    rel.Type,
		Animated: rel.Type == "AFFECTS" || rel.Type == "DEPENDS_ON",
	}
}
