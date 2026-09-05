import { useQuery } from '@tanstack/react-query';
import { Network, Server, MapPin } from 'lucide-react';
import { schemaApi } from '@/api/queries';

export interface ClusterNodeInfo {
  address: string;
  dataCenter: string;
  rack: string;
  releaseVersion: string;
  tokenCount: number;
  local: boolean;
  status: string;
}

interface ClusterPanelProps {
  collapsed?: boolean;
}

export function ClusterPanel({ collapsed }: ClusterPanelProps) {
  const { data: nodes, isLoading } = useQuery({
    queryKey: ['cluster', 'info'],
    queryFn: schemaApi.getClusterInfo,
    staleTime: 60 * 1000,
  });

  if (collapsed || isLoading) {
    return null;
  }

  if (!nodes || nodes.length === 0) {
    return null;
  }

  const dcs = groupBy(nodes, (n) => n.dataCenter || 'unknown');
  const ringNodes = ringLayout(nodes.length);

  return (
    <div
      className="rounded-lg p-4 animate-fade-in"
      style={{
        background: 'var(--bg-elevated)',
        border: '1px solid var(--border-primary)',
      }}
    >
      <div
        className="flex items-center gap-2 text-xs font-mono font-bold tracking-wider uppercase mb-3"
        style={{ color: 'var(--text-tertiary)' }}
      >
        <Network className="w-3.5 h-3.5" />
        Cluster · {nodes.length} node{nodes.length === 1 ? '' : 's'}
      </div>

      <div className="flex items-center gap-4">
        <svg viewBox="0 0 120 120" className="w-28 h-28 flex-shrink-0">
          <circle
            cx="60"
            cy="60"
            r="50"
            fill="none"
            stroke="var(--border-primary)"
            strokeWidth="1.5"
          />
          <circle
            cx="60"
            cy="60"
            r="50"
            fill="none"
            stroke="var(--accent-primary)"
            strokeWidth="1"
            opacity="0.2"
          />
          {nodes.map((node, i) => {
            const pos = ringNodes[i] ?? { x: 60, y: 60 };
            const color = node.local ? 'var(--accent-primary)' : '#22c55e';
            return (
              <g key={node.address}>
                <circle
                  cx={pos.x}
                  cy={pos.y}
                  r={node.local ? 7 : 5.5}
                  fill={color}
                  stroke={node.local ? 'var(--text-primary)' : 'none'}
                  strokeWidth={node.local ? 1.5 : 0}
                  opacity={0.9}
                />
                {node.local && (
                  <circle
                    cx={pos.x}
                    cy={pos.y}
                    r="11"
                    fill="none"
                    stroke="var(--accent-primary)"
                    strokeWidth="0.75"
                    opacity="0.4"
                  />
                )}
              </g>
            );
          })}
          <circle cx="60" cy="60" r="2" fill="var(--text-tertiary)" />
        </svg>

        <div className="flex-1 space-y-1.5 min-w-0">
          {Object.entries(dcs).map(([dc, dcNodes]) => (
            <div key={dc}>
              <div
                className="flex items-center gap-1.5 text-xs font-mono font-bold"
                style={{ color: 'var(--accent-primary)' }}
              >
                <MapPin className="w-3 h-3" />
                {dc}
                <span style={{ color: 'var(--text-tertiary)' }}>({dcNodes.length})</span>
              </div>
              {dcNodes.map((node) => (
                <div
                  key={node.address}
                  className="flex items-center justify-between gap-2 pl-4 text-xs font-mono"
                  style={{ color: 'var(--text-secondary)' }}
                >
                  <span className="flex items-center gap-1.5">
                    <Server className="w-3 h-3" style={{ color: node.local ? 'var(--accent-primary)' : '#22c55e' }} />
                    {node.address}
                  </span>
                  <span className="flex items-center gap-2" style={{ color: 'var(--text-tertiary)' }}>
                    <span>{node.rack}</span>
                    <span>v{node.releaseVersion}</span>
                    <span>{node.tokenCount} tokens</span>
                  </span>
                </div>
              ))}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function ringLayout(count: number): Array<{ x: number; y: number }> {
  const positions = [];
  for (let i = 0; i < count; i++) {
    const angle = (2 * Math.PI * i) / count - Math.PI / 2;
    positions.push({
      x: 60 + 50 * Math.cos(angle),
      y: 60 + 50 * Math.sin(angle),
    });
  }
  return positions;
}

function groupBy<T>(items: T[], key: (item: T) => string): Record<string, T[]> {
  const out: Record<string, T[]> = {};
  for (const item of items) {
    const k = key(item);
    if (!out[k]) out[k] = [];
    out[k].push(item);
  }
  return out;
}
