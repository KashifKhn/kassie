import { useState } from 'react';
import { Download, Loader2 } from 'lucide-react';
import { streamExport } from '@/api/export';
import { useToastStore } from '@/stores/toastStore';
import type { ApiError } from '@/api/types';

interface ExportButtonProps {
  keyspace: string;
  table: string;
  whereClause?: string;
}

export function ExportButton({ keyspace, table, whereClause }: ExportButtonProps) {
  const { success, error } = useToastStore();
  const [pending, setPending] = useState<'CSV' | 'JSON' | null>(null);

  const run = async (format: 'CSV' | 'JSON') => {
    setPending(format);
    try {
      const rows = await streamExport({ keyspace, table, whereClause, format });
      success(`Exported ${rows} rows to ${format}`);
    } catch (err) {
      const apiError = err as ApiError;
      error(apiError.message || 'Export failed');
    } finally {
      setPending(null);
    }
  };

  const busy = pending !== null;

  return (
    <div className="flex items-center gap-1">
      <ExportButtonItem
        label="CSV"
        pending={pending === 'CSV'}
        disabled={busy}
        onClick={() => run('CSV')}
      />
      <ExportButtonItem
        label="JSON"
        pending={pending === 'JSON'}
        disabled={busy}
        onClick={() => run('JSON')}
      />
    </div>
  );
}

function ExportButtonItem({
  label,
  pending,
  disabled,
  onClick,
}: {
  label: string;
  pending: boolean;
  disabled: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className="flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-mono rounded-md transition-all"
      style={{
        background: 'var(--accent-subtle)',
        color: 'var(--accent-primary)',
        cursor: disabled ? 'not-allowed' : 'pointer',
        opacity: disabled ? 0.6 : 1,
      }}
      onMouseEnter={(e) => {
        if (!disabled) {
          e.currentTarget.style.background = 'var(--accent-primary)';
          e.currentTarget.style.color = 'var(--text-inverse)';
        }
      }}
      onMouseLeave={(e) => {
        if (!disabled) {
          e.currentTarget.style.background = 'var(--accent-subtle)';
          e.currentTarget.style.color = 'var(--accent-primary)';
        }
      }}
      title={`Export current view as ${label}`}
    >
      {pending ? (
        <Loader2 className="w-3.5 h-3.5 animate-spin" />
      ) : (
        <Download className="w-3.5 h-3.5" />
      )}
      {label}
    </button>
  );
}
