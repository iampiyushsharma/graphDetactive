import dagre from "@dagrejs/dagre";

const dagreGraph = new dagre.graphlib.Graph();
dagreGraph.setDefaultEdgeLabel(() => ({}));

/**
 * Positions nodes automatically using dagre's hierarchical layout.
 * TB = Top to Bottom (best for root cause / top-down tracing)
 * LR = Left to Right (best for dependency chains)
 */
export const getLayoutedElements = (nodes: any[], edges: any[], direction = "TB") => {
	// Re-create the graph instance for every layout to prevent caching issues
	const graph = new dagre.graphlib.Graph();
	graph.setDefaultEdgeLabel(() => ({}));
	graph.setGraph({ rankdir: direction, nodesep: 80, ranksep: 120 });

	nodes.forEach((node) => {
		graph.setNode(node.id, { width: 220, height: 120 });
	});

	edges.forEach((edge) => {
		graph.setEdge(edge.source, edge.target);
	});

	dagre.layout(graph);

	const layoutedNodes = nodes.map((node) => {
		const nodeWithPosition = graph.node(node.id);
		return {
			...node,
			position: {
				// Offset to align from center
				x: nodeWithPosition.x - 110,
				y: nodeWithPosition.y - 60,
			},
		};
	});

	return { nodes: layoutedNodes, edges };
};
