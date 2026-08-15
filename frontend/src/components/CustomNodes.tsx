import React from "react";
import { Handle, Position } from "reactflow";
import { Server, Database as DbIcon, GitCommit, Sliders, HardDrive, ShieldAlert } from "lucide-react";

// Custom Incident Node
export const IncidentNode = ({ data }: any) => {
	const props = data.properties || {};
	return (
		<div className="px-4 py-3 shadow-xl rounded-lg bg-zinc-950 border border-red-500/50 text-white min-w-[200px] ring-2 ring-red-500/20">
			<Handle type="source" position={Position.Bottom} className="w-2 h-2 bg-red-500" />
			<div className="flex items-center gap-2 border-b border-red-500/20 pb-2 mb-2">
				<ShieldAlert className="w-5 h-5 text-red-500 animate-pulse" />
				<div>
					<div className="text-[10px] uppercase font-bold text-red-400">Incident</div>
					<div className="text-xs font-semibold text-red-100">{props.id || "INC"}</div>
				</div>
			</div>
			<div className="text-sm font-bold text-zinc-100 mb-1">{props.title}</div>
			<div className="flex items-center gap-1.5 mt-2">
				<span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-red-950 text-red-300 border border-red-800">
					{props.severity}
				</span>
				<span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-zinc-900 text-zinc-400 border border-zinc-700">
					{props.status}
				</span>
			</div>
		</div>
	);
};

// Custom Service Node
export const ServiceNode = ({ data }: any) => {
	const props = data.properties || {};
	const isCritical = props.status === "CRITICAL";
	const isDegraded = props.status === "DEGRADED";
	const borderClass = isCritical ? "border-red-500 ring-2 ring-red-500/25" : isDegraded ? "border-amber-500 ring-2 ring-amber-500/25" : "border-blue-500/50";
	const textClass = isCritical ? "text-red-400" : isDegraded ? "text-amber-400" : "text-blue-400";

	return (
		<div className={`px-4 py-3 shadow-xl rounded-lg bg-zinc-950 border text-white min-w-[200px] ${borderClass}`}>
			<Handle type="target" position={Position.Top} className="w-2 h-2 bg-blue-500" />
			<Handle type="source" position={Position.Bottom} className="w-2 h-2 bg-blue-500" />
			<div className="flex items-center gap-2 border-b border-zinc-800 pb-2 mb-2">
				<Server className={`w-5 h-5 ${textClass}`} />
				<div>
					<div className={`text-[10px] uppercase font-bold ${textClass}`}>Service</div>
					<div className="text-xs font-semibold text-zinc-300">{props.name}</div>
				</div>
			</div>
			<div className="text-xs text-zinc-400 flex justify-between">
				<span>Language:</span>
				<span className="font-semibold text-zinc-200">{props.language || "N/A"}</span>
			</div>
			<div className="flex items-center gap-1.5 mt-2 justify-between">
				<span className="text-[10px] text-zinc-500">Status:</span>
				<span className={`px-1.5 py-0.5 rounded text-[10px] font-bold ${
					isCritical ? "bg-red-950 text-red-300 border border-red-800" :
					isDegraded ? "bg-amber-950 text-amber-300 border border-amber-800" :
					"bg-green-950 text-green-300 border border-green-800"
				}`}>
					{props.status}
				</span>
			</div>
		</div>
	);
};

// Custom Database Node
export const DatabaseNode = ({ data }: any) => {
	const props = data.properties || {};
	const isDegraded = props.status === "DEGRADED";
	const borderClass = isDegraded ? "border-amber-500 ring-2 ring-amber-500/25" : "border-cyan-500/50";
	
	return (
		<div className={`px-4 py-3 shadow-xl rounded-lg bg-zinc-950 border text-white min-w-[200px] ${borderClass}`}>
			<Handle type="target" position={Position.Top} className="w-2 h-2 bg-cyan-500" />
			<Handle type="source" position={Position.Bottom} className="w-2 h-2 bg-cyan-500" />
			<div className="flex items-center gap-2 border-b border-zinc-800 pb-2 mb-2">
				<DbIcon className="w-5 h-5 text-cyan-400" />
				<div>
					<div className="text-[10px] uppercase font-bold text-cyan-400">Database</div>
					<div className="text-xs font-semibold text-zinc-300">{props.name}</div>
				</div>
			</div>
			<div className="text-xs text-zinc-400 flex justify-between">
				<span>Type:</span>
				<span className="font-semibold text-zinc-200">{props.type}</span>
			</div>
			<div className="flex items-center gap-1.5 mt-2 justify-between">
				<span className="text-[10px] text-zinc-500">Status:</span>
				<span className={`px-1.5 py-0.5 rounded text-[10px] font-bold ${
					isDegraded ? "bg-amber-950 text-amber-300 border border-amber-800" :
					"bg-green-950 text-green-300 border border-green-800"
				}`}>
					{props.status}
				</span>
			</div>
		</div>
	);
};

// Custom Deployment Node
export const DeploymentNode = ({ data }: any) => {
	const props = data.properties || {};
	return (
		<div className="px-4 py-3 shadow-xl rounded-lg bg-zinc-950 border border-purple-500/50 text-white min-w-[200px]">
			<Handle type="target" position={Position.Top} className="w-2 h-2 bg-purple-500" />
			<Handle type="source" position={Position.Bottom} className="w-2 h-2 bg-purple-500" />
			<div className="flex items-center gap-2 border-b border-zinc-800 pb-2 mb-2">
				<HardDrive className="w-5 h-5 text-purple-400" />
				<div>
					<div className="text-[10px] uppercase font-bold text-purple-400">Deployment</div>
					<div className="text-xs font-semibold text-purple-200">{props.id}</div>
				</div>
			</div>
			<div className="text-xs text-zinc-400 flex justify-between">
				<span>Version:</span>
				<span className="font-semibold text-zinc-200">{props.version}</span>
			</div>
			<div className="text-xs text-zinc-400 flex justify-between mt-1">
				<span>Env:</span>
				<span className="font-semibold text-zinc-200">{props.env}</span>
			</div>
		</div>
	);
};

// Custom Commit Node
export const CommitNode = ({ data }: any) => {
	const props = data.properties || {};
	return (
		<div className="px-4 py-3 shadow-xl rounded-lg bg-zinc-950 border border-orange-500/50 text-white min-w-[200px]">
			<Handle type="target" position={Position.Top} className="w-2 h-2 bg-orange-500" />
			<Handle type="source" position={Position.Bottom} className="w-2 h-2 bg-orange-500" />
			<div className="flex items-center gap-2 border-b border-zinc-800 pb-2 mb-2">
				<GitCommit className="w-5 h-5 text-orange-400" />
				<div>
					<div className="text-[10px] uppercase font-bold text-orange-400">Commit</div>
					<div className="text-xs font-semibold text-orange-200">{props.hash}</div>
				</div>
			</div>
			<div className="text-xs text-zinc-300 font-medium truncate mb-1" title={props.message}>
				"{props.message}"
			</div>
			<div className="text-[10px] text-zinc-500 mt-2">
				Author: <span className="text-zinc-400 font-semibold">{props.author}</span>
			</div>
		</div>
	);
};

// Custom ConfigChange Node
export const ConfigChangeNode = ({ data }: any) => {
	const props = data.properties || {};
	return (
		<div className="px-4 py-3 shadow-xl rounded-lg bg-zinc-950 border border-yellow-500/50 text-white min-w-[200px] ring-2 ring-yellow-500/10">
			<Handle type="target" position={Position.Top} className="w-2 h-2 bg-yellow-500" />
			<Handle type="source" position={Position.Bottom} className="w-2 h-2 bg-yellow-500" />
			<div className="flex items-center gap-2 border-b border-yellow-500/20 pb-2 mb-2">
				<Sliders className="w-5 h-5 text-yellow-400" />
				<div>
					<div className="text-[10px] uppercase font-bold text-yellow-400">Config Change</div>
					<div className="text-xs font-semibold text-yellow-200 truncate max-w-[130px]" title={props.key}>
						{props.key}
					</div>
				</div>
			</div>
			<div className="flex items-center justify-between text-xs mt-1">
				<span className="text-zinc-500">Old:</span>
				<span className="text-red-400 font-semibold bg-red-950/40 px-1 rounded line-through">{props.oldValue}</span>
			</div>
			<div className="flex items-center justify-between text-xs mt-1">
				<span className="text-zinc-500">New:</span>
				<span className="text-green-400 font-semibold bg-green-950/40 px-1 rounded">{props.newValue}</span>
			</div>
		</div>
	);
};
