import { useQuery } from '@tanstack/react-query';
import { Gauge, RefreshCw, Loader2 } from 'lucide-react';
import { dataApi } from '@/api/queries';

export interface TraceEvent {
  activity: string;
  source: string;
  elapsedUs: number;
  thread: string;
}

export interface TraceData {
  events: TraceEvent[];
  durationUs: number;
  coordinator: string;
  ready: boolean;
}

interface TracePanelProps {
  traceId: string;
}

export function TracePanel({ traceId }: TracePanelProps) {
  const { data: trace, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['trace', traceId],
    queryFn: () => dataApi.getTrace(traceId),
    refetchInterval: (query) => (query.state.data?.ready ? false : 2000),
    enabled: Boolean(traceId),
  });

  if (!traceId) {
    return null;
  }

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-xs font-mono px-3 py-2" style={{ color: 'var(--text-tertiary)' }}>
        <Loader2 className="w-3.5 h-3.5 animate-spin" />
        Loading trace…
      </div>
    );
  }

  if (!trace?.ready) {
    return (
      <div
        className="flex items-center justify-between gap-2 text-xs font-mono px-3 py-2 rounded-lg"
        style={{ background: 'var(--bg-primary)', border: '1px solid var(--border-primary)', color: 'var(--text-tertiary)' }}
      >
        <span>Trace {traceId.slice(0, 8)}… not ready yet — Cassandra writes traces asynchronously</span>
        <button onClick={() => refetch()} className="flex items-center gap-1" style={{ color: 'var(--accent-primary)' }}>
          <RefreshCw className={`w-3 h-3 ${isFetching ? 'animate-spin' : ''}`} />
          Retry
        </button>
      </div>
    );
  }

  const totalUs = trace.durationUs > 0 ? trace.durationUs : maxElapsed(trace.events);

  return (
    <div
      className="rounded-lg overflow-hidden"
      style={{ border: '1px solid var(--border-primary)' }}
    >
      <div
        className="flex items-center justify-between px-3 py-2 text-xs font-mono"
        style={{ background: 'var(--bg-tertiary)', color: 'var(--text-secondary)' }}
      >
        <span className="flex items-center gap-2">
          <Gauge className="w-3.5 h-3.5" style={{ color: 'var(--accent-primary)' }} />
          Trace — {formatUs(totalUs)} on {trace.coordinator || 'coordinator'}
        </span>
        <button onClick={() => refetch()} style={{ color: 'var(--text-tertiary)' }}>
          <RefreshCw className={`w-3 h-3 ${isFetching ? 'animate-spin' : ''}`} />
        </button>
      </div>

      <div className="max-h-64 overflow-auto">
        {trace.events.map((e, i) => (
          <div
            key={i}
            className="flex items-center gap-2 px-3 py-1 text-xs font-mono"
            style={{ borderBottom: '1px solid var(--border-primary)' }}
          >
            <span className="w-24 flex-shrink-0" style={{ color: 'var(--accent-primary)' }}>
              +{formatUs(e.elapsedUs)}
            </span>
            <div className="flex-1 min-w-0">
              <div
                className="h-3 rounded-sm"
                style={{
                  width: `${Math.max(2, (e.elapsedUs / totalUs) * 100)}%`,
                  background: 'var(--accent-primary)',
                  opacity: 0.3 + 0.7 * Math.min(1, e.elapsedUs / totalUs),
                  minWidth: '4px',
                }}
              />
            </div>
            <span className="flex-2 truncate" style={{ color: 'var(--text-primary)' }} title={e.activity}>
              {e.activity}
            </span>
            <span className="w-32 flex-shrink-0 truncate text-right" style={{ color: 'var(--text-tertiary)' }}>
              {e.source}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function formatUs(us: number): string {
  if (us <= 0) return '0';
  if (us < 1000) return `${us}µs`;
  if (us < 1_000_000) return `${(us / 1000).toFixed(1)}ms`;
  return `${(us / 1_000_000).toFixed(2)}s`;
}

function maxElapsed(events: TraceEvent[]): number {
  let max = 0;
  for (const e of events) {
    if (e.elapsedUs > max) max = e.elapsedUs;
  }
  return max || 1;
}
