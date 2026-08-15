import React, { useEffect } from "react";
import ReactFlow, {
	Background,
	Controls,
	MiniMap,
	BackgroundVariant,
	useNodesState,
	useEdgesState,
} from "reactflow";
import "reactflow/dist/style.css";
import {
	IncidentNode,
	ServiceNode,
	DatabaseNode,
	DeploymentNode,
	CommitNode,
	ConfigChangeNode,
} from "./CustomNodes";
import { getLayoutedElements } from "../utils/layout";

// Map Neo4j node labels to Custom React Flow components
const nodeTypes = {
	Incident: IncidentNode,
	Service: ServiceNode,
	Database: DatabaseNode,
	Deployment: DeploymentNode,
	Commit: CommitNode,
	ConfigChange: ConfigChangeNode,
};

interface GraphExplorerProps {
	nodes: any[];
	edges: any[];
	onNodeClick: (node: any) => void;
	direction?: "TB" | "LR";
}

export default function GraphExplorer({
	nodes: initialNodes,
	edges: initialEdges,
	onNodeClick,
	direction = "TB",
}: GraphExplorerProps) {
	const [nodes, setNodes, onNodesChange] = useNodesState([]);
	const [edges, setEdges, onEdgesChange] = useEdgesState([]);

	useEffect(() => {
		if (initialNodes.length === 0) {
			setNodes([]);
			setEdges([]);
			return;
		}

		// Calculate automatic node positioning
		const layouted = getLayoutedElements(initialNodes, initialEdges, direction);
		setNodes(layouted.nodes);
		setEdges(layouted.edges);
	}, [initialNodes, initialEdges, direction, setNodes, setEdges]);

	const handleNodeClick = (_event: React.MouseEvent, node: any) => {
		onNodeClick(node);
	};

	return (
		<div className="w-full h-full bg-zinc-950 rounded-xl overflow-hidden border border-zinc-800 relative">
			<ReactFlow
				nodes={nodes}
				edges={edges}
				nodeTypes={nodeTypes}
				onNodesChange={onNodesChange}
				onEdgesChange={onEdgesChange}
				onNodeClick={handleNodeClick}
				fitView
				minZoom={0.2}
				maxZoom={1.5}
			>
				<Background color="#3f3f46" gap={16} size={1} variant={BackgroundVariant.Dots} />
				<Controls className="bg-zinc-900 border border-zinc-850 text-white rounded fill-white" />
				<MiniMap
					nodeColor={(n) => {
						if (n.type === "Incident") return "#ef4444";
						if (n.type === "Service") return "#3b82f6";
						if (n.type === "Database") return "#06b6d4";
						if (n.type === "Deployment") return "#a855f7";
						if (n.type === "Commit") return "#f97316";
						if (n.type === "ConfigChange") return "#eab308";
						return "#71717a";
					}}
					maskColor="rgba(0, 0, 0, 0.6)"
					className="bg-zinc-900 border border-zinc-800 rounded overflow-hidden"
				/>
			</ReactFlow>
		</div>
	);
}
