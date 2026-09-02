import { useQuery } from '@tanstack/react-query';
import { BarChart3, Hash, HardDrive } from 'lucide-react';
import { statsApi, queryKeys } from '@/api/queries';

interface TableStatsCardProps {
  keyspace: string;
  table: string;
}

function formatBytes(bytes: number): string {
  if (bytes <= 0) return '—';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  return `${value.toFixed(value < 10 && unit > 0 ? 1 : 0)} ${units[unit]}`;
}

function formatCount(count: number): string {
  if (count <= 0) return '—';
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1)}M`;
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}k`;
  return count.toString();
}

export function TableStatsCard({ keyspace, table }: TableStatsCardProps) {
  const { data: stats, isLoading, isError } = useQuery({
    queryKey: queryKeys.stats(keyspace, table),
    queryFn: () => statsApi.getTableStats(keyspace, table),
    staleTime: 60 * 1000,
  });

  if (isError) {
    return null;
  }

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
        <BarChart3 className="w-3.5 h-3.5" />
        Table Stats
      </div>

      {isLoading || !stats ? (
        <p className="text-xs font-mono" style={{ color: 'var(--text-tertiary)' }}>
          Loading stats...
        </p>
      ) : (
        <div className="space-y-2">
          <StatRow
            icon={<Hash className="w-3.5 h-3.5" />}
            label="Rows"
            value={formatCount(stats.rowCount)}
            hint={stats.estimateAvailable ? 'estimate' : 'exact count'}
          />
          <StatRow
            icon={<HardDrive className="w-3.5 h-3.5" />}
            label="Avg partition"
            value={formatBytes(stats.meanPartitionSizeBytes)}
          />
          <StatRow
            icon={<HardDrive className="w-3.5 h-3.5" />}
            label="Max partition"
            value={formatBytes(stats.maxPartitionSizeBytes)}
          />
        </div>
      )}
    </div>
  );
}

function StatRow({
  icon,
  label,
  value,
  hint,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  hint?: string;
}) {
  return (
    <div className="flex items-center justify-between text-xs font-mono">
      <span
        className="flex items-center gap-1.5"
        style={{ color: 'var(--text-secondary)' }}
      >
        {icon}
        {label}
      </span>
      <span className="flex items-center gap-2">
        <span style={{ color: 'var(--text-primary)' }}>{value}</span>
        {hint && (
          <span style={{ color: 'var(--text-tertiary)', fontStyle: 'italic' }}>
            {hint}
          </span>
        )}
      </span>
    </div>
  );
}
