"use client";

import React, { useState, useEffect } from "react";
import {
	Activity,
	AlertOctagon,
	Play,
	Radio,
	Zap,
	Info,
	Server,
	Database,
	GitCommit,
	Sliders,
	HardDrive,
	ShieldAlert,
	RefreshCw,
	Eye
} from "lucide-react";
import GraphExplorer from "../components/GraphExplorer";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export default function Home() {
	const [dbStatus, setDbStatus] = useState<"connected" | "disconnected" | "checking">("checking");
	const [incidents, setIncidents] = useState<any[]>([]);
	const [selectedIncident, setSelectedIncident] = useState<any | null>(null);
	const [activeGraph, setActiveGraph] = useState<{ nodes: any[]; edges: any[] }>({ nodes: [], edges: [] });
	const [selectedNode, setSelectedNode] = useState<any | null>(null);
	const [graphType, setGraphType] = useState<"root-cause" | "blast-radius" | "dependencies" | "none">("none");
	const [isLoading, setIsLoading] = useState(false);
	const [errorMsg, setErrorMsg] = useState<string | null>(null);

	// Check Database Connection
	const checkStatus = async () => {
		try {
			const res = await fetch(`${API_BASE}/api/status`);
			if (!res.ok) throw new Error("Status service degraded");
			const data = await res.json();
			setDbStatus(data.database === "connected" ? "connected" : "disconnected");
		} catch (err) {
			setDbStatus("disconnected");
		}
	};

	// Load Incidents
	const loadIncidents = async () => {
		try {
			setIsLoading(true);
			setErrorMsg(null);
			const res = await fetch(`${API_BASE}/api/incidents`);
			if (!res.ok) {
				if (res.status === 503) {
					setDbStatus("disconnected");
					throw new Error("CognoDB is offline. Please configure credentials in the backend environment.");
				}
				throw new Error("Failed to load incidents");
			}
			const data = await res.json();
			setIncidents(data || []);
		} catch (err: any) {
			setErrorMsg(err.message);
		} finally {
			setIsLoading(false);
		}
	};

	const refreshAll = () => {
		checkStatus();
		loadIncidents();
		if (selectedIncident) {
			setSelectedIncident(null);
			setActiveGraph({ nodes: [], edges: [] });
			setSelectedNode(null);
			setGraphType("none");
		}
	};

	useEffect(() => {
		checkStatus();
		loadIncidents();
	}, []);

	// Select Incident
	const handleSelectIncident = (incident: any) => {
		setSelectedIncident(incident);
		setActiveGraph({ nodes: [], edges: [] });
		setSelectedNode(null);
		setGraphType("none");
	};

	// Trace Root Cause
	const handleTraceRootCause = async () => {
		if (!selectedIncident) return;
		try {
			setIsLoading(true);
			setErrorMsg(null);
			setGraphType("root-cause");
			setSelectedNode(null);
			const res = await fetch(`${API_BASE}/api/incidents/${selectedIncident.id}/root-cause`);
			if (!res.ok) throw new Error("Failed to compute root cause path");
			const data = await res.json();
			setActiveGraph(data);
		} catch (err: any) {
			setErrorMsg(err.message);
		} finally {
			setIsLoading(false);
		}
	};

	// Analyze Blast Radius
	const handleAnalyzeBlastRadius = async () => {
		if (!selectedIncident) return;
		try {
			setIsLoading(true);
			setErrorMsg(null);
			setGraphType("blast-radius");
			setSelectedNode(null);
			// Target ID: resolve map ID
			const targetId = selectedIncident.serviceName === "Gateway" ? "gateway" : 
							 selectedIncident.serviceName === "AuthService" ? "auth-service" :
							 selectedIncident.serviceName === "ProductCatalogService" ? "catalog-service" :
							 selectedIncident.serviceName === "CartService" ? "cart-service" :
							 selectedIncident.serviceName === "OrderService" ? "order-service" :
							 selectedIncident.serviceName === "PaymentService" ? "payment-service" : "order-service";

			const res = await fetch(`${API_BASE}/api/services/${targetId}/blast-radius`);
			if (!res.ok) throw new Error("Failed to compute blast radius");
			const data = await res.json();
			setActiveGraph(data);
		} catch (err: any) {
			setErrorMsg(err.message);
		} finally {
			setIsLoading(false);
		}
	};

	// Show Dependencies
	const handleShowDependencies = async () => {
		if (!selectedIncident) return;
		try {
			setIsLoading(true);
			setErrorMsg(null);
			setGraphType("dependencies");
			setSelectedNode(null);
			const targetId = selectedIncident.serviceName === "Gateway" ? "gateway" : 
							 selectedIncident.serviceName === "AuthService" ? "auth-service" :
							 selectedIncident.serviceName === "ProductCatalogService" ? "catalog-service" :
							 selectedIncident.serviceName === "CartService" ? "cart-service" :
							 selectedIncident.serviceName === "OrderService" ? "order-service" :
							 selectedIncident.serviceName === "PaymentService" ? "payment-service" : "order-service";

			const res = await fetch(`${API_BASE}/api/services/${targetId}/dependencies`);
			if (!res.ok) throw new Error("Failed to retrieve service dependencies");
			const data = await res.json();
			setActiveGraph(data);
		} catch (err: any) {
			setErrorMsg(err.message);
		} finally {
			setIsLoading(false);
		}
	};

	const handleNodeClick = (node: any) => {
		setSelectedNode(node);
	};

	return (
		<div className="flex flex-col min-h-screen bg-zinc-950 text-zinc-100 font-sans">
			{/* Navbar Header */}
			<header className="flex items-center justify-between px-6 py-4 border-b border-zinc-800 bg-zinc-900/50 backdrop-blur">
				<div className="flex items-center gap-3">
					<div className="p-2 rounded-lg bg-red-500/10 border border-red-500/20 text-red-500">
						<Activity className="w-6 h-6 animate-pulse" />
					</div>
					<div>
						<h1 className="text-xl font-bold tracking-tight bg-gradient-to-r from-zinc-100 via-zinc-200 to-zinc-400 bg-clip-text text-transparent">
							Graph Detective
						</h1>
						<p className="text-[10px] font-semibold tracking-wide text-zinc-500 uppercase">
							Production Root Cause & Blast Radius Platform
						</p>
					</div>
				</div>

				<div className="flex items-center gap-4">
					{/* Health Indicator Badge */}
					<div className={`flex items-center gap-2 px-3 py-1.5 rounded-full border text-xs font-semibold ${
						dbStatus === "connected" ? "bg-green-950/40 text-green-400 border-green-800" :
						dbStatus === "disconnected" ? "bg-red-950/40 text-red-400 border-red-800" :
						"bg-zinc-900 text-zinc-400 border-zinc-800"
					}`}>
						<span className={`w-2 h-2 rounded-full ${
							dbStatus === "connected" ? "bg-green-400 animate-pulse" :
							dbStatus === "disconnected" ? "bg-red-400 animate-pulse" :
							"bg-zinc-600"
						}`} />
						CognoDB: {dbStatus === "connected" ? "Connected" : dbStatus === "disconnected" ? "Offline" : "Checking..."}
					</div>

					<button 
						onClick={refreshAll}
						className="p-2 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 rounded-lg border border-zinc-800 transition"
						title="Reset & Refresh Dashboard"
					>
						<RefreshCw className="w-4 h-4" />
					</button>
				</div>
			</header>

			{/* Main Workspace Layout */}
			<main className="flex-1 flex overflow-hidden">
				{/* Sidebar Section */}
				<aside className="w-80 border-r border-zinc-800 bg-zinc-900/20 flex flex-col overflow-y-auto">
					<div className="p-4 border-b border-zinc-800">
						<h2 className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-3">
							Incidents Catalog
						</h2>

						{/* Database Down Banner */}
						{dbStatus === "disconnected" && (
							<div className="p-3 mb-4 rounded-lg bg-red-950/20 border border-red-500/20 text-red-400 text-xs">
								<div className="flex gap-2">
									<AlertOctagon className="w-4 h-4 shrink-0 mt-0.5" />
									<div>
										<span className="font-bold">CognoDB Offline</span>
										<p className="mt-1 text-zinc-400">Please check environment configs (`COGNO_DB_URI` and `COGNO_DB_PASSWORD`) and start the backend service.</p>
									</div>
								</div>
							</div>
						)}

						{/* Skeletons/Loading */}
						{isLoading && incidents.length === 0 ? (
							<div className="space-y-2">
								{[1, 2, 3].map((i) => (
									<div key={i} className="h-20 w-full bg-zinc-900 border border-zinc-800 rounded-lg animate-pulse" />
								))}
							</div>
						) : incidents.length === 0 ? (
							<div className="text-center py-6 text-zinc-500 text-xs border border-zinc-800 border-dashed rounded-lg">
								No active incidents found. Check seeder script.
							</div>
						) : (
							<div className="space-y-2">
								{incidents.map((inc) => {
									const isSelected = selectedIncident?.id === inc.id;
									return (
										<button
											key={inc.id}
											onClick={() => handleSelectIncident(inc)}
											className={`w-full text-left p-3 rounded-lg border transition ${
												isSelected 
													? "bg-zinc-900 border-zinc-600 ring-1 ring-zinc-650"
													: "bg-zinc-900/50 border-zinc-850 hover:bg-zinc-900 hover:border-zinc-800"
											}`}
										>
											<div className="flex items-center justify-between mb-1.5">
												<span className={`px-1.5 py-0.5 rounded text-[9px] font-bold ${
													inc.severity === "CRITICAL" ? "bg-red-950 text-red-300 border border-red-800/40" :
													inc.severity === "HIGH" ? "bg-amber-950 text-amber-300 border border-amber-800/40" :
													"bg-blue-950 text-blue-300 border border-blue-800/40"
												}`}>
													{inc.severity}
												</span>
												<span className="text-[10px] text-zinc-500">
													{new Date(inc.createdAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
												</span>
											</div>
											<div className="text-xs font-bold text-zinc-250 mb-1">{inc.title}</div>
											<div className="text-[10px] text-zinc-500">Target: <span className="text-zinc-400 font-semibold">{inc.serviceName}</span></div>
										</button>
									);
								})}
							</div>
						)}
					</div>

					{/* Controller Actions Panel */}
					{selectedIncident && (
						<div className="p-4 flex-1 flex flex-col bg-zinc-900/40">
							<div className="pb-4 border-b border-zinc-800">
								<h3 className="text-xs font-bold text-zinc-400 mb-2">Incident Details</h3>
								<div className="text-xs font-semibold text-zinc-200 mb-1.5">{selectedIncident.title}</div>
								<div className="text-[11px] text-zinc-400 leading-relaxed bg-zinc-950 p-2.5 rounded border border-zinc-850">
									{selectedIncident.description}
								</div>
							</div>

							<div className="mt-4 space-y-2">
								<h3 className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2">
									Investigation Actions
								</h3>

								<button
									onClick={handleTraceRootCause}
									disabled={isLoading}
									className={`w-full flex items-center justify-between px-3 py-2 text-xs font-semibold rounded-lg border transition ${
										graphType === "root-cause"
											? "bg-red-950/40 text-red-300 border-red-800"
											: "bg-zinc-900 hover:bg-zinc-800 border-zinc-850 text-zinc-300"
									}`}
								>
									<span className="flex items-center gap-2">
										<Zap className="w-4 h-4 text-red-500 shrink-0" />
										Trace Root Cause
									</span>
									<span className="px-1 text-[9px] bg-zinc-800 text-zinc-400 rounded font-mono">4 Hops</span>
								</button>

								<button
									onClick={handleAnalyzeBlastRadius}
									disabled={isLoading}
									className={`w-full flex items-center justify-between px-3 py-2 text-xs font-semibold rounded-lg border transition ${
										graphType === "blast-radius"
											? "bg-cyan-950/40 text-cyan-300 border-cyan-800"
											: "bg-zinc-900 hover:bg-zinc-800 border-zinc-850 text-zinc-300"
									}`}
								>
									<span className="flex items-center gap-2">
										<Eye className="w-4 h-4 text-cyan-400 shrink-0" />
										Analyze Blast Radius
									</span>
									<span className="px-1 text-[9px] bg-zinc-800 text-zinc-400 rounded font-mono">Transitive</span>
								</button>

								<button
									onClick={handleShowDependencies}
									disabled={isLoading}
									className={`w-full flex items-center justify-between px-3 py-2 text-xs font-semibold rounded-lg border transition ${
										graphType === "dependencies"
											? "bg-blue-950/40 text-blue-300 border-blue-800"
											: "bg-zinc-900 hover:bg-zinc-800 border-zinc-850 text-zinc-300"
									}`}
								>
									<span className="flex items-center gap-2">
										<Server className="w-4 h-4 text-blue-500 shrink-0" />
										Show Dependencies
									</span>
									<span className="px-1 text-[9px] bg-zinc-800 text-zinc-400 rounded font-mono">1 Hop</span>
								</button>
							</div>
						</div>
					)}
				</aside>

				{/* Visual Graph Canvas Area */}
				<section className="flex-1 flex flex-col p-6 bg-zinc-950 relative">
					{errorMsg && (
						<div className="absolute top-6 left-6 right-6 z-10 p-3 rounded-lg bg-red-950/40 border border-red-500/30 text-red-400 text-xs flex justify-between items-center backdrop-blur">
							<span>Error: {errorMsg}</span>
							<button onClick={() => setErrorMsg(null)} className="font-bold underline hover:text-white">Dismiss</button>
						</div>
					)}

					{/* Loading Overlay */}
					{isLoading && (
						<div className="absolute inset-6 bg-zinc-950/70 z-10 flex flex-col items-center justify-center rounded-xl backdrop-blur-sm">
							<RefreshCw className="w-8 h-8 text-zinc-400 animate-spin mb-3" />
							<span className="text-sm font-semibold text-zinc-300">Executing OpenCypher Query...</span>
							<span className="text-[10px] text-zinc-500 mt-1 font-mono">Traversing CognoDB index-free adjacency</span>
						</div>
					)}

					<div className="flex-1 min-h-0 flex gap-6">
						{/* Canvas Viewport */}
						<div className="flex-1 flex flex-col">
							<div className="flex items-center justify-between mb-3">
								<div className="text-xs font-bold text-zinc-500 uppercase tracking-wider">
									Interactive Graph Visualizer {graphType !== "none" && `(${graphType})`}
								</div>
								{graphType !== "none" && (
									<div className="text-[10px] text-zinc-400">
										Tip: Drag and zoom canvas. Click nodes to inspect properties.
									</div>
								)}
							</div>

							{activeGraph && activeGraph.nodes && activeGraph.nodes.length > 0 ? (
								<GraphExplorer
									nodes={activeGraph.nodes || []}
									edges={activeGraph.edges || []}
									onNodeClick={handleNodeClick}
									direction={graphType === "root-cause" ? "TB" : "LR"}
								/>
							) : (
								<div className="flex-1 border border-zinc-800 border-dashed rounded-xl flex flex-col items-center justify-center bg-zinc-900/10 p-8 text-center">
									<Radio className="w-12 h-12 text-zinc-700 animate-pulse mb-4" />
									<h3 className="text-sm font-bold text-zinc-400 mb-1">Investigation Map Ready</h3>
									<p className="text-xs text-zinc-500 max-w-sm">
										Select an incident from the catalog on the left and choose an action to build the dependency map.
									</p>
								</div>
							)}
						</div>

						{/* Node details panel */}
						{selectedNode && (
							<div className="w-80 border border-zinc-800 rounded-xl bg-zinc-900/40 p-4 flex flex-col overflow-y-auto animate-fade-in shrink-0">
								<div className="flex items-center gap-2 pb-3 border-b border-zinc-800 mb-4">
									{selectedNode.type === "Incident" && <ShieldAlert className="w-5 h-5 text-red-500" />}
									{selectedNode.type === "Service" && <Server className="w-5 h-5 text-blue-500" />}
									{selectedNode.type === "Database" && <Database className="w-5 h-5 text-cyan-500" />}
									{selectedNode.type === "Deployment" && <HardDrive className="w-5 h-5 text-purple-500" />}
									{selectedNode.type === "Commit" && <GitCommit className="w-5 h-5 text-orange-500" />}
									{selectedNode.type === "ConfigChange" && <Sliders className="w-5 h-5 text-yellow-500" />}
									<div>
										<h4 className="text-xs font-bold text-zinc-400 uppercase tracking-wider">{selectedNode.type}</h4>
										<div className="text-xs font-semibold text-zinc-200 truncate max-w-[190px]">{selectedNode.data.label}</div>
									</div>
								</div>

								<div className="flex-1 space-y-4">
									<div>
										<h5 className="text-[10px] font-bold text-zinc-500 uppercase tracking-wider mb-2">Properties</h5>
										<div className="space-y-2 bg-zinc-950 p-3 rounded-lg border border-zinc-850 font-mono text-[11px] overflow-x-auto">
											{Object.entries(selectedNode.data.properties || {}).map(([key, val]: any) => {
												if (typeof val === "object" && val !== null) {
													return (
														<div key={key} className="py-1 border-b border-zinc-900/50 last:border-0">
															<span className="text-zinc-500">{key}:</span>
															<pre className="text-zinc-300 mt-1 whitespace-pre-wrap">{JSON.stringify(val, null, 2)}</pre>
														</div>
													);
												}
												return (
													<div key={key} className="flex justify-between py-1 border-b border-zinc-900/50 last:border-0 gap-4">
														<span className="text-zinc-500 shrink-0">{key}:</span>
														<span className="text-zinc-300 text-right break-all">{String(val)}</span>
													</div>
												);
											})}
										</div>
									</div>

									{selectedNode.type === "ConfigChange" && (
										<div className="p-3 bg-yellow-950/20 border border-yellow-500/20 rounded-lg text-yellow-400 text-xs">
											<div className="flex gap-2">
												<Info className="w-4 h-4 shrink-0 mt-0.5" />
												<div>
													<span className="font-bold">Configuration Culprit</span>
													<p className="mt-1 text-zinc-400 leading-relaxed">
														This parameter was updated right before the outage. Traveled via deployment commits.
													</p>
												</div>
											</div>
										</div>
									)}
								</div>
							</div>
						)}
					</div>
				</section>
			</main>
		</div>
	);
}
