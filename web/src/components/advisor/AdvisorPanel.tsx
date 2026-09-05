import { useQuery } from '@tanstack/react-query';
import { AlertTriangle, Info, Wrench } from 'lucide-react';
import { schemaApi, queryKeys } from '@/api/queries';

interface AdvisorPanelProps {
  keyspace: string;
}

const severityStyle: Record<string, { color: string; icon: React.ReactNode }> = {
  warning: { color: 'var(--warning)', icon: <AlertTriangle className="w-3.5 h-3.5" /> },
  info: { color: 'var(--info)', icon: <Info className="w-3.5 h-3.5" /> },
};

export function AdvisorPanel({ keyspace }: AdvisorPanelProps) {
  const { data, isLoading } = useQuery({
    queryKey: queryKeys.advisor(keyspace),
    queryFn: () => schemaApi.analyzeKeyspace(keyspace),
    staleTime: 5 * 60 * 1000,
    enabled: Boolean(keyspace),
  });

  if (isLoading || !data) {
    return null;
  }

  if (data.findings.length === 0) {
    return (
      <div
        className="rounded-lg p-4 animate-fade-in"
        style={{
          background: 'var(--bg-elevated)',
          border: '1px solid var(--border-primary)',
        }}
      >
        <div
          className="flex items-center gap-2 text-xs font-mono font-bold tracking-wider uppercase mb-2"
          style={{ color: 'var(--text-tertiary)' }}
        >
          <Wrench className="w-3.5 h-3.5" />
          Advisor
        </div>
        <p className="text-xs font-mono" style={{ color: '#22c55e' }}>
          ✓ No findings across {data.tablesAnalyzed} tables
        </p>
      </div>
    );
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
        className="flex items-center justify-between text-xs font-mono font-bold tracking-wider uppercase mb-3"
        style={{ color: 'var(--text-tertiary)' }}
      >
        <span className="flex items-center gap-2">
          <Wrench className="w-3.5 h-3.5" />
          Advisor
        </span>
        <span>
          {data.findings.length} finding{data.findings.length === 1 ? '' : 's'} · {data.tablesAnalyzed} tables
        </span>
      </div>

      <div className="space-y-2">
        {data.findings.map((f, i) => {
          const style = severityStyle[f.severity] ?? {
            color: 'var(--info)',
            icon: <Info className="w-3.5 h-3.5" />,
          };
          return (
            <div
              key={f.rule + f.table + i}
              className="rounded-lg p-3 space-y-1.5"
              style={{ background: 'var(--bg-primary)', border: '1px solid var(--border-primary)' }}
            >
              <div className="flex items-center gap-2 text-xs font-mono" style={{ color: style.color }}>
                {style.icon}
                <span className="uppercase font-bold">{f.severity}</span>
                <span style={{ color: 'var(--text-tertiary)' }}>· {f.rule}</span>
                {f.table && (
                  <span style={{ color: 'var(--text-primary)' }}>{f.table}</span>
                )}
              </div>
              <p className="text-xs font-mono" style={{ color: 'var(--text-secondary)', lineHeight: 1.5 }}>
                {f.message}
              </p>
              {f.remediation && (
                <pre
                  className="text-[10px] font-mono p-2 rounded overflow-auto whitespace-pre-wrap"
                  style={{ background: 'var(--bg-secondary)', color: 'var(--accent-primary)' }}
                >
                  {f.remediation}
                </pre>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
